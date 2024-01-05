/*
Package workers provides index, sync, and metadata optimization background workers.

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package workers

import (
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
)

var log = event.Log
var stop = make(chan bool, 1)

// Start runs the metadata, share & sync background workers at regular intervals.
func Start(conf *config.Config) {
	interval := conf.WakeupInterval()

	// Disabled in safe mode?
	if interval.Seconds() <= 0 {
		log.Warnf("config: disabled metadata, share & sync background workers")
		return
	}

	ticker := time.NewTicker(interval)

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
				RunMeta(conf)
				RunShare(conf)
				RunSync(conf)
			}
		}
	}()
}

// Stop shuts down all service workers.
func Stop() {
	stop <- true
}

// RunMeta runs the metadata worker once.
func RunMeta(conf *config.Config) {
	if !mutex.IndexWorkersRunning() {
		go func() {
			worker := NewMeta(conf)

			delay := time.Minute
			interval := entity.MetadataUpdateInterval

			if err := worker.Start(delay, interval, false); err != nil {
				log.Warnf("metadata: %s", err)
			}
		}()
	}
}

// RunShare runs the share worker once.
func RunShare(conf *config.Config) {
	if !mutex.ShareWorker.Running() {
		go func() {
			worker := NewShare(conf)
			if err := worker.Start(); err != nil {
				log.Warnf("share: %s", err)
			}
		}()
	}
}

// RunSync runs the sync worker once.
func RunSync(conf *config.Config) {
	if !mutex.SyncWorker.Running() {
		go func() {
			worker := NewSync(conf)
			if err := worker.Start(); err != nil {
				log.Warnf("sync: %s", err)
			}
		}()
	}
}
