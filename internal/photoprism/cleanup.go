package photoprism

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/fastwalk"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
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

	if err := mutex.MainWorker.Start(); err != nil {
		log.Warnf("cleanup: %s (start)", err.Error())
		return thumbs, orphans, err
	}

	defer mutex.MainWorker.Stop()

	if opt.Dry {
		log.Infof("cleanup: dry run, nothing will actually be removed")
	}

	// Find and remove orphan thumbnail files.
	hashes, err := query.FileHashes()

	if err != nil {
		return thumbs, orphans, err
	}

	thumbPath := w.conf.ThumbPath()

	if err := fastwalk.Walk(thumbPath, func(fileName string, info os.FileMode) error {
		base := filepath.Base(fileName)

		if info.IsDir() || strings.HasPrefix(base, ".") {
			return nil
		}

		i := strings.Index(base, "_")

		if i < 39 {
			return nil
		}

		hash := base[:i]
		logName := txt.Quote(fs.RelName(fileName, thumbPath))

		if ok := hashes[hash]; ok {
			// Do nothing.
		} else if opt.Dry {
			thumbs++
			log.Debugf("cleanup: orphan thumbnail %s would be removed", logName)
		} else if err := os.Remove(fileName); err != nil {
			log.Warnf("cleanup: %s in %s", err, logName)
		} else {
			thumbs++
			log.Debugf("cleanup: removed orphan thumbnail %s", logName)
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
			log.Infof("cleanup: orphan photo %s would be removed", txt.Quote(p.PhotoUID))
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

	if err := query.PurgeOrphans(); err != nil {
		log.Errorf("cleanup: %s (purge orphans)", err)
	}

	// Update counts and views if needed.
	if len(deleted) > 0 {
		if err := entity.UpdatePhotoCounts(); err != nil {
			log.Errorf("cleanup: %s", err)
		}

		event.EntitiesDeleted("photos", deleted)
	}

	return thumbs, orphans, nil
}

// Cancel stops the current operation.
func (w *CleanUp) Cancel() {
	mutex.MainWorker.Cancel()
}
