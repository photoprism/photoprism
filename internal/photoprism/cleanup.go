package photoprism

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

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
func (w *CleanUp) Start(opt CleanUpOptions) (thumbs int, orphans int, sidecars int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cleanup: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	originalsPath := w.conf.OriginalsPath()

	// Check if originals folder is empty.
	if fs.DirIsEmpty(originalsPath) {
		return thumbs, orphans, sidecars, err
	}

	if err = mutex.MainWorker.Start(); err != nil {
		log.Warnf("cleanup: %s (start)", err)
		return thumbs, orphans, sidecars, err
	}

	defer mutex.MainWorker.Stop()

	if opt.Dry {
		log.Infof("cleanup: dry run, nothing will actually be removed")
	}

	// Find and remove orphan photo index entries.
	cleanupStart := time.Now()
	photos, err := query.OrphanPhotos()

	if err != nil {
		return thumbs, orphans, sidecars, err
	}

	var deleted []string

	// Delete orphaned photos from the index and remaining sidecar files from storage, if any.
	for _, p := range photos {
		if mutex.MainWorker.Canceled() {
			return thumbs, orphans, sidecars, errors.New("cleanup canceled")
		}

		if opt.Dry {
			orphans++
			log.Infof("cleanup: %s would be removed from index", p.String())
			continue
		}

		// Deletes the index entry and remaining sidecar files outside the "originals" folder.
		if n, err := DeletePhoto(p, true, false); err != nil {
			sidecars += n
			log.Errorf("cleanup: %s (remove orphans)", err)
		} else {
			orphans++
			sidecars += n
			deleted = append(deleted, p.PhotoUID)
			log.Infof("cleanup: removed %s from index", p.String())
		}
	}

	log.Infof("cleanup: removed %s and %s [%s]", english.Plural(orphans, "index entry", "index entries"), english.Plural(sidecars, "sidecar file", "sidecar files"), time.Since(cleanupStart))

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
		if err = query.PurgeOrphans(); err != nil {
			log.Errorf("index: %s (purge orphans)", err)
		}
	}

	// Remove thumbnail files.
	thumbs, err = w.Thumbs(opt)

	// Only update counts if anything was deleted.
	if len(deleted) > 0 {
		// Update precalculated photo and file counts.
		if err = entity.UpdateCounts(); err != nil {
			log.Warnf("index: %s (update counts)", err)
		}

		// Update album, subject, and label cover thumbs.
		if err = query.UpdateCovers(); err != nil {
			log.Warnf("index: %s (update covers)", err)
		}

		// Show success notification.
		event.EntitiesDeleted("photos", deleted)
	}

	return thumbs, orphans, sidecars, err
}

// Thumbs removes orphan thumbnail files.
func (w *CleanUp) Thumbs(opt CleanUpOptions) (thumbs int, err error) {
	cleanupStart := time.Now()

	var fileHashes, thumbHashes query.HashMap

	// Fetch existing media and thumb file hashes.
	if fileHashes, err = query.FileHashMap(); err != nil {
		return thumbs, err
	} else if thumbHashes, err = query.ThumbHashMap(); err != nil {
		return thumbs, err
	}

	// At least one SHA1 checksum found?
	if len(fileHashes) == 0 {
		log.Info("cleanup: empty index, aborting search for orphaned thumbnails")
		return thumbs, err
	}

	// Thumbnails storage path.
	thumbPath := w.conf.ThumbCachePath()

	log.Info("cleanup: searching for orphaned thumbnails")

	// Find and remove orphan thumbnail files.
	err = fastwalk.Walk(thumbPath, func(fileName string, info os.FileMode) error {
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
		} else if ok = thumbHashes[hash]; ok {
			// Do nothing.
		} else if opt.Dry {
			thumbs++
			log.Debugf("cleanup: thumbnail %s would be removed", logName)
		} else if err := os.Remove(fileName); err != nil {
			log.Warnf("cleanup: %s in %s", err, logName)
		} else {
			thumbs++
			log.Debugf("cleanup: removed thumbnail %s from cache", logName)
		}

		return nil
	})

	log.Infof("cleanup: removed %s [%s]", english.Plural(thumbs, "thumbnail file", "thumbnail files"), time.Since(cleanupStart))

	return thumbs, err
}

// Cancel stops the current operation.
func (w *CleanUp) Cancel() {
	mutex.MainWorker.Cancel()
}
