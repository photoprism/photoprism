package workers

import (
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
)

var log = event.Log
var stop = make(chan bool, 1)

// Start runs PhotoPrism background workers every wakeup interval.
func Start(conf *config.Config) {
	ticker := time.NewTicker(conf.WakeupInterval())

	go func() {
		for {
			select {
			case <-stop:
				log.Info("shutting down workers")
				ticker.Stop()
				mutex.Share.Cancel()
				mutex.Sync.Cancel()
				return
			case <-ticker.C:
				StartShare(conf)
				StartSync(conf)
			}
		}
	}()
}

// Stop shuts down all service workers.
func Stop() {
	stop <- true
}

// StartShare runs the share worker once.
func StartShare(conf *config.Config) {
	if !mutex.Share.Busy() {
		go func() {
			s := NewShare(conf)
			if err := s.Start(); err != nil {
				log.Error(err)
			}
		}()
	}
}

// StartShare runs the sync worker once.
func StartSync(conf *config.Config) {
	if !mutex.Sync.Busy() {
		go func() {
			s := NewSync(conf)
			if err := s.Start(); err != nil {
				log.Error(err)
			}
		}()
	}
}
