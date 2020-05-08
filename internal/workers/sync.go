package workers

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/remote"
)

// Sync represents a sync worker.
type Sync struct {
	conf *config.Config
}

// NewSync returns a new service sync worker.
func NewSync(conf *config.Config) *Sync {
	return &Sync{
		conf: conf,
	}
}

// Start starts the sync worker.
func (s *Sync) Start() (err error) {
	if err := mutex.Sync.Start(); err != nil {
		event.Error(fmt.Sprintf("sync: %s", err.Error()))
		return err
	}

	defer mutex.Sync.Stop()

	f := form.AccountSearch{
		Sync: true,
	}

	accounts, err := query.Accounts(f)

	for _, a := range accounts {
		if a.AccType != remote.ServiceWebDAV {
			continue
		}

		if a.AccErrors > a.RetryLimit {
			log.Warnf("sync: %s failed more than %d times", a.AccName, a.RetryLimit)
			continue
		}

		// Values updated in account: AccError, AccErrors, SyncStatus, SyncDate
		accError := a.AccError
		accErrors := a.AccErrors
		syncStatus := a.SyncStatus
		syncDate := a.SyncDate
		synced := false

		switch a.SyncStatus {
		case entity.AccountSyncStatusRefresh:
			if complete, err := s.refresh(a); err != nil {
				accErrors++
				accError = err.Error()
			} else if complete {
				accErrors = 0
				accError = ""

				if a.SyncDownload {
					syncStatus = entity.AccountSyncStatusDownload
				} else if a.SyncUpload {
					syncStatus = entity.AccountSyncStatusUpload
				} else {
					syncStatus = entity.AccountSyncStatusSynced
					syncDate.Time = time.Now()
					syncDate.Valid = true
				}
			}
		case entity.AccountSyncStatusDownload:
			if complete, err := s.download(a); err != nil {
				accErrors++
				accError = err.Error()
			} else if complete {
				if a.SyncUpload {
					syncStatus = entity.AccountSyncStatusUpload
				} else {
					synced = true
					syncStatus = entity.AccountSyncStatusSynced
					syncDate.Time = time.Now()
					syncDate.Valid = true
				}
			}
		case entity.AccountSyncStatusUpload:
			if complete, err := s.upload(a); err != nil {
				accErrors++
				accError = err.Error()
			} else if complete {
				synced = true
				syncStatus = entity.AccountSyncStatusSynced
				syncDate.Time = time.Now()
				syncDate.Valid = true
			}
		case entity.AccountSyncStatusSynced:
			if a.SyncDate.Valid && a.SyncDate.Time.Before(time.Now().Add(time.Duration(-1*a.SyncInterval)*time.Second)) {
				syncStatus = entity.AccountSyncStatusRefresh
			}
		default:
			syncStatus = entity.AccountSyncStatusRefresh
		}

		if mutex.Sync.Canceled() {
			return nil
		}

		if err := entity.Db().First(&a, a.ID).Error; err != nil {
			log.Errorf("sync: %s", err.Error())
			return err
		}

		// Only update the following fields to avoid overwriting other settings
		a.AccError = accError
		a.AccErrors = accErrors
		a.SyncStatus = syncStatus
		a.SyncDate = syncDate

		if err := entity.Db().Save(&a).Error; err != nil {
			log.Errorf("sync: %s", err.Error())
		} else if synced {
			event.Publish("sync.synced", event.Data{"account": a})
		}
	}

	return err
}
