package workers

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
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
		if a.AccType != service.WebDAV {
			continue
		}

		// Failed too often? The stored row decides, since an admin may have changed it during this run,
		// and the account is skipped either way, so its loaded copy is never written back.
		if a.RetryLimit > 0 && a.AccErrors > a.RetryLimit {
			if res := entity.Db().Model(&entity.Service{}).
				Where("id = ? AND acc_sync = 1 AND retry_limit > 0 AND acc_errors > retry_limit", a.ID).
				UpdateColumn("acc_sync", false); res.Error != nil {
				w.logErr(res.Error)
			} else if res.RowsAffected > 0 {
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
				accError = clean.ErrorBytes(err, txt.ClipError)
			} else if complete {
				accErrors = 0
				accError = ""

				switch {
				case a.SyncDownload:
					syncStatus = entity.SyncStatusDownload
				case a.SyncUpload:
					syncStatus = entity.SyncStatusUpload
				default:
					syncStatus = entity.SyncStatusSynced
					syncDate.Time = time.Now()
					syncDate.Valid = true
				}
			}
		case entity.SyncStatusDownload:
			if complete, downloadErr := w.download(a); downloadErr != nil {
				accErrors++
				accError = clean.ErrorBytes(downloadErr, txt.ClipError)
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
			if complete, uploadErr := w.upload(a); uploadErr != nil {
				accErrors++
				accError = clean.ErrorBytes(uploadErr, txt.ClipError)
				syncStatus = entity.SyncStatusRefresh
			} else if complete {
				synced = true
				syncStatus = entity.SyncStatusSynced
				syncDate.Time = time.Now()
				syncDate.Valid = true
			}
		case entity.SyncStatusSynced:
			if syncDue(a, time.Now()) {
				syncStatus = entity.SyncStatusRefresh
			}
		default:
			syncStatus = entity.SyncStatusRefresh
		}

		if mutex.SyncWorker.Canceled() {
			return nil
		}

		// Only update the following fields to avoid overwriting other settings
		if updateErr := a.Updates(entity.Values{
			"AccError":   accError,
			"AccErrors":  accErrors,
			"SyncStatus": syncStatus,
			"SyncDate":   syncDate}); updateErr != nil {
			w.logErr(updateErr)
		} else if synced {
			event.Publish("sync.synced", event.Data{"account": a})
		}
	}

	return err
}

// syncDue reports whether a synced account is due for a refresh, comparing whole seconds so no stored interval overflows.
func syncDue(a entity.Service, now time.Time) bool {
	return a.SyncInterval > 0 && a.SyncDate.Valid && a.SyncDate.Time.Unix() < now.Unix()-int64(a.SyncInterval)
}
