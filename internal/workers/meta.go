package workers

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
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

// originalsPath returns the original media files path as string.
func (worker *Meta) originalsPath() string {
	return worker.conf.OriginalsPath()
}

// Start starts the metadata worker.
func (worker *Meta) Start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("meta-worker: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err := mutex.MetaWorker.Start(); err != nil {
		log.Warnf("meta-worker: %s (start)", err.Error())
		return err
	}

	defer mutex.MetaWorker.Stop()

	log.Debugf("meta-worker: starting routine check")

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
				return errors.New("meta-worker: check canceled")
			}

			if done[photo.PhotoUID] {
				continue
			}

			done[photo.PhotoUID] = true

			if updated, err := photo.Optimize(); err != nil {
				log.Errorf("meta-worker: %s (optimize photo)", err)
			} else if updated {
				optimized++
				log.Debugf("meta-worker: optimized photo %s", photo.String())
			}
		}

		if mutex.MetaWorker.Canceled() {
			return errors.New("meta-worker: check canceled")
		}

		offset += limit

		time.Sleep(100 * time.Millisecond)
	}

	if optimized > 0 {
		log.Infof("meta-worker: optimized %d photos", optimized)
	}

	if err := query.ResetPhotoQuality(); err != nil {
		log.Warnf("meta-worker: %s (reset photo quality)", err.Error())
	}

	if err := entity.UpdatePhotoCounts(); err != nil {
		log.Warnf("meta-worker: %s (update photo counts)", err.Error())
	}

	moments := photoprism.NewMoments(worker.conf)

	if err := moments.Start(); err != nil {
		log.Error(err)
	}

	runtime.GC()

	return nil
}
