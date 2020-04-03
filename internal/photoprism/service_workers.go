package photoprism

import (
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
)

func ServiceWorkers(conf *config.Config) chan bool {
	ticker := time.NewTicker(5 * time.Minute)
	stop := make(chan bool, 1)

	go func() {
		for {
			select {
			case <-stop:
				log.Info("stopping service workers")
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

	return stop
}
