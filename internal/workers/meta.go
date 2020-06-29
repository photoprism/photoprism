package workers

import (
	"errors"
	"runtime"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
)

// Meta represents a background metadata maintenance worker.
type Meta struct {
	conf *config.Config
}

// NewMeta returns a new background metadata maintenance worker.
func NewMeta(conf *config.Config) *Meta {
	return &Meta{conf: conf}
}

// logError logs an error message if err is not nil.
func (worker *Meta) logError(err error) {
	if err != nil {
		log.Errorf("metadata: %s", err.Error())
	}
}

// logWarn logs a warning message if err is not nil.
func (worker *Meta) logWarn(err error) {
	if err != nil {
		log.Warnf("metadata: %s", err.Error())
	}
}

// originalsPath returns the original media files path as string.
func (worker *Meta) originalsPath() string {
	return worker.conf.OriginalsPath()
}

// Start starts the metadata worker.
func (worker *Meta) Start() (err error) {
	if err := mutex.MetaWorker.Start(); err != nil {
		worker.logWarn(err)
		return err
	}

	defer func() {
		mutex.MetaWorker.Stop()

		if err := recover(); err != nil {
			log.Errorf("metadata: %s (worker panic)", err)
		}
	}()

	log.Debugf("metadata: starting routine check")

	done := make(map[string]bool)

	limit := 50
	offset := 0
	optimized := 0

	for {
		photos, err := query.PhotosCheck(limit, offset)

		if err != nil {
			return err
		}

		if len(photos) == 0 {
			break
		} else if offset == 0 {

		}

		for _, photo := range photos {
			if mutex.MetaWorker.Canceled() {
				return errors.New("metadata: check canceled")
			}

			if done[photo.PhotoUID] {
				continue
			}

			done[photo.PhotoUID] = true

			if updated, err := photo.Optimize(); err != nil {
				log.Errorf("metadata: %s", err)
			} else if updated {
				optimized++
				log.Debugf("metadata: optimized photo %s", photo.String())
			}
		}

		if mutex.MetaWorker.Canceled() {
			return errors.New("metadata: check canceled")
		}

		offset += limit

		time.Sleep(100 * time.Millisecond)
	}

	if optimized > 0 {
		log.Infof("metadata: optimized %d photos", optimized)
	}

	worker.logError(query.ResetPhotoQuality())

	worker.logError(entity.UpdatePhotoCounts())

	moments := photoprism.NewMoments(worker.conf)

	if err := moments.Start(); err != nil {
		log.Error(err)
	}

	runtime.GC()

	return nil
}
