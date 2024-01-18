package workers

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/remote"
	"github.com/photoprism/photoprism/internal/search"
)

// Sync represents a sync worker.
type Sync struct {
	conf *config.Config
}

// NewSync returns a new sync worker.
func NewSync(conf *config.Config) *Sync {
	return &Sync{
		conf: conf,
	}
}

// logErr logs an error message if err is not nil.
func (w *Sync) logErr(err error) {
	if err != nil {
		log.Errorf("sync: %s", err.Error())
	}
}

// logWarn logs a warning message if err is not nil.
func (w *Sync) logWarn(err error) {
	if err != nil {
		log.Warnf("sync: %s", err.Error())
	}
}

// Start starts the sync worker.
func (w *Sync) Start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("sync: %s (worker panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err := mutex.SyncWorker.Start(); err != nil {
		return err
	}

	defer mutex.SyncWorker.Stop()

	f := form.SearchServices{
		Sync: true,
	}

	accounts, err := search.Accounts(f)

	for _, a := range accounts {
		if a.AccType != remote.ServiceWebDAV {
			continue
		}

		// Failed too often?
		if a.RetryLimit > 0 && a.AccErrors > a.RetryLimit {
			a.AccSync = false

			if err := entity.Db().Save(&a).Error; err != nil {
				w.logErr(err)
			} else {
				log.Warnf("sync: disabled sync, %s failed more than %d times", a.AccName, a.RetryLimit)
			}

			continue
		}

		// Values updated in account: AccError, AccErrors, SyncStatus, SyncDate
		accError := a.AccError
		accErrors := a.AccErrors
		syncStatus := a.SyncStatus
		syncDate := a.SyncDate
		synced := false

		switch a.SyncStatus {
		case entity.SyncStatusRefresh:
			if complete, err := w.refresh(a); err != nil {
				accErrors++
				accError = err.Error()
			} else if complete {
				accErrors = 0
				accError = ""

				if a.SyncDownload {
					syncStatus = entity.SyncStatusDownload
				} else if a.SyncUpload {
					syncStatus = entity.SyncStatusUpload
				} else {
					syncStatus = entity.SyncStatusSynced
					syncDate.Time = time.Now()
					syncDate.Valid = true
				}
			}
		case entity.SyncStatusDownload:
			if complete, err := w.download(a); err != nil {
				accErrors++
				accError = err.Error()
				syncStatus = entity.SyncStatusRefresh
			} else if complete {
				if a.SyncUpload {
					syncStatus = entity.SyncStatusUpload
				} else {
					synced = true
					syncStatus = entity.SyncStatusSynced
					syncDate.Time = time.Now()
					syncDate.Valid = true
				}
			}
		case entity.SyncStatusUpload:
			if complete, err := w.upload(a); err != nil {
				accErrors++
				accError = err.Error()
				syncStatus = entity.SyncStatusRefresh
			} else if complete {
				synced = true
				syncStatus = entity.SyncStatusSynced
				syncDate.Time = time.Now()
				syncDate.Valid = true
			}
		case entity.SyncStatusSynced:
			if a.SyncDate.Valid && a.SyncDate.Time.Before(time.Now().Add(time.Duration(-1*a.SyncInterval)*time.Second)) {
				syncStatus = entity.SyncStatusRefresh
			}
		default:
			syncStatus = entity.SyncStatusRefresh
		}

		if mutex.SyncWorker.Canceled() {
			return nil
		}

		// Only update the following fields to avoid overwriting other settings
		if err := a.Updates(map[string]interface{}{
			"AccError":   accError,
			"AccErrors":  accErrors,
			"SyncStatus": syncStatus,
			"SyncDate":   syncDate}); err != nil {
			w.logErr(err)
		} else if synced {
			event.Publish("sync.synced", event.Data{"account": a})
		}
	}

	return err
}
