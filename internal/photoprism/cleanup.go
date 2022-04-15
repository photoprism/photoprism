package photoprism

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fastwalk"
	"github.com/photoprism/photoprism/pkg/fs"
)

// CleanUp represents a worker that deletes unneeded data and files.
type CleanUp struct {
	conf *config.Config
}

// NewCleanUp returns a new cleanup worker.
func NewCleanUp(conf *config.Config) *CleanUp {
	instance := &CleanUp{
		conf: conf,
	}

	return instance
}

// Start removes orphan index entries and thumbnails.
func (w *CleanUp) Start(opt CleanUpOptions) (thumbs int, orphans int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cleanup: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err = mutex.MainWorker.Start(); err != nil {
		log.Warnf("cleanup: %s (start)", err.Error())
		return thumbs, orphans, err
	}

	defer mutex.MainWorker.Stop()

	if opt.Dry {
		log.Infof("cleanup: dry run, nothing will actually be removed")
	}

	var fileHashes, thumbHashes query.HashMap

	// Fetch existing media and thumb file hashes.
	if fileHashes, err = query.FileHashMap(); err != nil {
		return thumbs, orphans, err
	} else if thumbHashes, err = query.ThumbHashMap(); err != nil {
		return thumbs, orphans, err
	}

	// Thumbnails storage path.
	thumbPath := w.conf.ThumbPath()

	// Find and remove orphan thumbnail files.
	if err := fastwalk.Walk(thumbPath, func(fileName string, info os.FileMode) error {
		base := filepath.Base(fileName)

		if info.IsDir() || strings.HasPrefix(base, ".") {
			return nil
		}

		// Example: 01244519acf35c62a5fea7a5a7dcefdbec4fb2f5_3x3_resize.png
		i := strings.Index(base, "_")

		if i < 39 {
			return nil
		}

		hash := base[:i]
		logName := clean.Log(fs.RelName(fileName, thumbPath))

		if ok := fileHashes[hash]; ok {
			// Do nothing.
		} else if ok := thumbHashes[hash]; ok {
			// Do nothing.
		} else if opt.Dry {
			thumbs++
			log.Debugf("cleanup: thumbnail %s would be removed", logName)
		} else if err := os.Remove(fileName); err != nil {
			log.Warnf("cleanup: %s in %s", err, logName)
		} else {
			thumbs++
			log.Debugf("cleanup: removed thumbnail %s", logName)
		}

		return nil
	}); err != nil {
		return thumbs, orphans, err
	}

	// Find and remove orphan photo index entries.
	photos, err := query.OrphanPhotos()

	if err != nil {
		return thumbs, orphans, err
	}

	var deleted []string

	for _, p := range photos {
		if mutex.MainWorker.Canceled() {
			return thumbs, orphans, errors.New("cleanup canceled")
		}

		if opt.Dry {
			orphans++
			log.Infof("cleanup: orphan photo %s would be removed", clean.Log(p.PhotoUID))
			continue
		}

		if err := Delete(p); err != nil {
			log.Errorf("cleanup: %s (remove orphan photo)", err.Error())
		} else {
			orphans++
			deleted = append(deleted, p.PhotoUID)
			log.Debugf("cleanup: removed orphan photo %s", p.PhotoUID)
		}
	}

	// Remove orphan index entries.
	if opt.Dry {
		if files, err := query.OrphanFiles(); err != nil {
			log.Errorf("index: %s (find orphan files)", err)
		} else if l := len(files); l > 0 {
			log.Infof("index: found %s", english.Plural(l, "orphan file", "orphan files"))
		} else {
			log.Infof("index: found no orphan files")
		}
	} else {
		if err := query.PurgeOrphans(); err != nil {
			log.Errorf("index: %s (purge orphans)", err)
		}
	}

	// Only update counts if anything was deleted.
	if len(deleted) > 0 {
		// Update precalculated photo and file counts.
		if err := entity.UpdateCounts(); err != nil {
			log.Warnf("index: %s (update counts)", err)
		}

		// Update album, subject, and label cover thumbs.
		if err := query.UpdateCovers(); err != nil {
			log.Warnf("index: %s (update covers)", err)
		}

		// Show success notification.
		event.EntitiesDeleted("photos", deleted)
	}

	return thumbs, orphans, nil
}

// Cancel stops the current operation.
func (w *CleanUp) Cancel() {
	mutex.MainWorker.Cancel()
}
