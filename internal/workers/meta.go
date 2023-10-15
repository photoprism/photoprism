package workers

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
)

// Meta represents a background index and metadata optimization worker.
type Meta struct {
	conf    *config.Config
	lastRun time.Time
}

// NewMeta returns a new Meta worker.
func NewMeta(conf *config.Config) *Meta {
	return &Meta{conf: conf}
}

// originalsPath returns the original media files path as string.
func (w *Meta) originalsPath() string {
	return w.conf.OriginalsPath()
}

// Start metadata optimization routine.
func (w *Meta) Start(delay, interval time.Duration, force bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("index: %s (worker panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err = mutex.MetaWorker.Start(); err != nil {
		return err
	}

	defer mutex.MetaWorker.Stop()

	// Check time when worker was last executed.
	updateIndex := force || w.lastRun.Before(time.Now().Add(-1*entity.IndexUpdateInterval))

	// Run faces worker if needed.
	if updateIndex || entity.UpdateFaces.Load() {
		log.Debugf("index: running face recognition")
		if faces := photoprism.NewFaces(w.conf); faces.Disabled() {
			log.Debugf("index: skipping face recognition")
		} else if err := faces.Start(photoprism.FacesOptions{}); err != nil {
			log.Warn(err)
		}
	}

	// Refresh index metadata.
	log.Debugf("index: updating metadata")

	start := time.Now()
	settings := w.conf.Settings()
	done := make(map[string]bool)
	limit := 1000
	offset := 0
	optimized := 0

	for {
		photos, err := query.PhotosMetadataUpdate(limit, offset, delay, interval)

		if err != nil {
			return err
		}

		if len(photos) == 0 {
			break
		}

		for _, photo := range photos {
			if mutex.MetaWorker.Canceled() {
				return errors.New("index: metadata optimization canceled")
			}

			if done[photo.PhotoUID] {
				continue
			}

			done[photo.PhotoUID] = true

			updated, merged, err := photo.Optimize(settings.StackMeta(), settings.StackUUID(), settings.Features.Estimates, force)

			if err != nil {
				log.Errorf("index: %s in optimization worker", err)
			} else if updated {
				optimized++
				log.Debugf("index: updated photo %s", photo.String())
			}

			for _, p := range merged {
				log.Infof("index: merged %s", p.PhotoUID)
				done[p.PhotoUID] = true
			}
		}

		if mutex.MetaWorker.Canceled() {
			return errors.New("index: optimization canceled")
		}

		offset += limit
	}

	if optimized > 0 {
		log.Infof("index: updated %s [%s]", english.Plural(optimized, "photo", "photos"), time.Since(start))
		updateIndex = true
	}

	// Only update index if necessary.
	if updateIndex {
		// Set photo quality scores to -1 if files are missing.
		if err = query.FlagHiddenPhotos(); err != nil {
			log.Warnf("index: %s in optimization worker", err)
		}

		// Run moments worker.
		if moments := photoprism.NewMoments(w.conf); moments == nil {
			log.Errorf("index: failed updating moments")
		} else if err = moments.Start(); err != nil {
			log.Warnf("moments: %s in optimization worker", err)
		}

		// Update precalculated photo and file counts.
		if err = entity.UpdateCounts(); err != nil {
			log.Warnf("index: %s in optimization worker", err)
		}

		// Update album, subject, and label cover thumbs.
		if err = query.UpdateCovers(); err != nil {
			log.Warnf("index: %s in optimization worker", err)
		}
	}

	// Update time when worker was last executed.
	w.lastRun = entity.TimeStamp()

	// Run garbage collection.
	runtime.GC()

	return nil
}
