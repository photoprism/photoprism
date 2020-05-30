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

// Prism represents a background maintenance worker.
type Prism struct {
	conf *config.Config
}

// NewPrism returns a new background maintenance worker.
func NewPrism(conf *config.Config) *Prism {
	return &Prism{conf: conf}
}

// logError logs an error message if err is not nil.
func (worker *Prism) logError(err error) {
	if err != nil {
		log.Errorf("prism: %s", err.Error())
	}
}

// logWarn logs a warning message if err is not nil.
func (worker *Prism) logWarn(err error) {
	if err != nil {
		log.Warnf("prism: %s", err.Error())
	}
}

// originalsPath returns the original media files path as string.
func (worker *Prism) originalsPath() string {
	return worker.conf.OriginalsPath()
}

// Start starts the prism worker.
func (worker *Prism) Start() (err error) {
	if err := mutex.PrismWorker.Start(); err != nil {
		worker.logWarn(err)
		return err
	}

	defer func() {
		mutex.PrismWorker.Stop()

		if err := recover(); err != nil {
			log.Errorf("prism: %s [panic]", err)
		}
	}()

	done := make(map[string]bool)

	limit := 50
	offset := 0

	for {
		photos, err := query.PhotosMaintenance(limit, offset)

		if err != nil {
			return err
		}

		if len(photos) == 0 {
			break
		} else if offset == 0 {
			log.Infof("prism: starting metadata optimization")
		}

		for _, photo := range photos {
			if mutex.PrismWorker.Canceled() {
				return errors.New("prism: optimization canceled")
			}

			if done[photo.PhotoUID] {
				continue
			}

			done[photo.PhotoUID] = true

			worker.logError(photo.Maintain())
		}

		if mutex.PrismWorker.Canceled() {
			return errors.New("prism: optimization canceled")
		}

		offset += limit

		time.Sleep(100 * time.Millisecond)
	}

	if len(done) > 0 {
		log.Infof("prism: optimized %d photos", len(done))
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
