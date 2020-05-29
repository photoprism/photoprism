package photoprism

import (
	"fmt"
	"runtime"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
)

// Moments represents a worker that creates albums based on popular locations, dates and categories.
type Moments struct {
	conf *config.Config
}

// NewMoments returns a new purge worker.
func NewMoments(conf *config.Config) *Moments {
	instance := &Moments{
		conf: conf,
	}

	return instance
}

// Start creates albums based on popular locations, dates and categories.
func (m *Moments) Start() (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("moments: %s [panic]", err)
		}
	}()

	if err := mutex.MainWorker.Start(); err != nil {
		err = fmt.Errorf("moments: %s", err.Error())
		event.Error(err.Error())
		return err
	}

	defer func() {
		mutex.MainWorker.Stop()

		if err := recover(); err != nil {
			log.Errorf("moments: %s [panic]", err)
		} else {
			runtime.GC()
		}
	}()

	return nil
}

// Cancel stops the current operation.
func (m *Moments) Cancel() {
	mutex.MainWorker.Cancel()
}
