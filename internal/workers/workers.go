package workers

import (
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
)

var log = event.Log
var stop = make(chan bool, 1)

func Start(conf *config.Config) {
	ticker := time.NewTicker(5 * time.Minute) // TODO

	go func() {
		for {
			select {
			case <-stop:
				log.Info("shutting down service workers")
				ticker.Stop()
				mutex.Share.Cancel()
				mutex.Sync.Cancel()
				return
			case <-ticker.C:
				if !mutex.Share.Busy() {
					go func() {
						// Start
						s := NewShare(conf)
						if err := s.Start(); err != nil {
							log.Error(err)
						}
					}()
				}

				if !mutex.Sync.Busy() {
					go func() {
						// Start
						s := NewSync(conf)
						if err := s.Start(); err != nil {
							log.Error(err)
						}
					}()
				}
			}
		}
	}()
}

func Stop() {
	stop <- true
}
