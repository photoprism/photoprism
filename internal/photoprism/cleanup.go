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
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fastwalk"
	"github.com/photoprism/photoprism/pkg/fs"
)

// CleanUp represents a worker that deletes orphaned index entries, sidecar files and thumbnails.
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

// Start index and cache cleanup.
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

	if err = mutex.IndexWorker.Start(); err != nil {
		log.Warnf("cleanup: %s (start)", err)
		return thumbs, orphans, sidecars, err
	}

	defer mutex.IndexWorker.Stop()

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
		if mutex.IndexWorker.Canceled() {
			return thumbs, orphans, sidecars, errors.New("cleanup canceled")
		}

		if opt.Dry {
			orphans++
			log.Infof("cleanup: %s would be removed from index", p.String())
			continue
		}

		// Deletes the index entry and remaining sidecar files outside the "originals" folder.
		if n, deleteErr := DeletePhoto(&p, true, false); deleteErr != nil {
			sidecars += n
			log.Errorf("cleanup: %s (remove orphans)", deleteErr)
		} else {
			orphans++
			sidecars += n
			deleted = append(deleted, p.PhotoUID)
			log.Infof("cleanup: removed %s from index", p.String())
		}
	}

	log.Infof("cleanup: removed %s and deleted %s [%s]", english.Plural(orphans, "index entry", "index entries"), english.Plural(sidecars, "sidecar file", "sidecar files"), time.Since(cleanupStart))

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

	// Remove orphaned media and thumbnail cache files.
	thumbs, err = w.Cache(opt)

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

// Cache removes orphaned media and thumbnail cache files.
func (w *CleanUp) Cache(opt CleanUpOptions) (deleted int, err error) {
	cleanupStart := time.Now()

	var fileHashes, thumbHashes query.HashMap

	// Fetch existing media and thumb file hashes.
	if fileHashes, err = query.FileHashMap(); err != nil {
		return deleted, err
	} else if thumbHashes, err = query.ThumbHashMap(); err != nil {
		return deleted, err
	}

	// At least one SHA1 checksum found?
	if len(fileHashes) == 0 {
		log.Info("cleanup: empty index, aborting search for orphaned cache files")
		return deleted, err
	}

	// Cache directories.
	dirs := []string{w.conf.MediaCachePath(), w.conf.ThumbCachePath()}

	log.Info("cleanup: searching for orphaned cache files")

	// Find and delete orphaned thumbnail files.
	for _, dir := range dirs {
		err = fastwalk.Walk(dir, func(fileName string, info os.FileMode) error {
			base := filepath.Base(fileName)

			if info.IsDir() || strings.HasPrefix(base, ".") {
				return nil
			}

			// Example: 01244519acf35c62a5fea7a5a7dcefdbec4fb2f5_3x3_resize.png
			i := strings.IndexAny(base, "_.")

			if i < 39 {
				return nil
			}

			hash := base[:i]
			logName := clean.Log(fs.RelName(fileName, filepath.Dir(dir)))

			if ok := fileHashes[hash]; ok {
				// Do nothing.
			} else if ok = thumbHashes[hash]; ok {
				// Do nothing.
			} else if opt.Dry {
				deleted++
				log.Debugf("cleanup: %s would be deleted", logName)
			} else if err := os.Remove(fileName); err != nil {
				log.Warnf("cleanup: %s in %s", err, logName)
			} else {
				deleted++
				log.Debugf("cleanup: deleted %s from cache", logName)
			}

			return nil
		})
	}

	log.Infof("cleanup: deleted %s from cache [%s]", english.Plural(deleted, "file", "files"), time.Since(cleanupStart))

	return deleted, err
}

// Cancel stops the current operation.
func (w *CleanUp) Cancel() {
	mutex.IndexWorker.Cancel()
}
