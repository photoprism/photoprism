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
				mutex.MetaWorker.Cancel()
				mutex.ShareWorker.Cancel()
				mutex.SyncWorker.Cancel()
				return
			case <-ticker.C:
				StartMeta(conf)
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

// StartMeta runs the metadata worker once.
func StartMeta(conf *config.Config) {
	if !mutex.WorkersBusy() {
		go func() {
			worker := NewMeta(conf)
			if err := worker.Start(); err != nil {
				log.Error(err)
			}
		}()
	}
}

// StartShare runs the share worker once.
func StartShare(conf *config.Config) {
	if !mutex.ShareWorker.Busy() {
		go func() {
			worker := NewShare(conf)
			if err := worker.Start(); err != nil {
				log.Error(err)
			}
		}()
	}
}

// StartShare runs the sync worker once.
func StartSync(conf *config.Config) {
	if !mutex.SyncWorker.Busy() {
		go func() {
			worker := NewSync(conf)
			if err := worker.Start(); err != nil {
				log.Error(err)
			}
		}()
	}
}
