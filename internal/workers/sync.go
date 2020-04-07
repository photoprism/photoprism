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
	q    *query.Query
}

// NewSync returns a new service sync worker.
func NewSync(conf *config.Config) *Sync {
	return &Sync{
		conf: conf,
		q:    query.New(conf.Db()),
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

	db := s.conf.Db()
	q := s.q

	accounts, err := q.Accounts(f)

	for _, a := range accounts {
		if a.AccType != remote.ServiceWebDAV {
			continue
		}

		if a.AccErrors > a.RetryLimit {
			log.Warnf("sync: %s failed more than %d times", a.AccName, a.RetryLimit)
			continue
		}

		switch a.SyncStatus {
		case entity.AccountSyncStatusRefresh:
			if complete, err := s.refresh(a); err != nil {
				a.AccErrors++
				a.AccError = err.Error()
			} else if complete {
				a.AccErrors = 0
				a.AccError = ""

				if a.SyncDownload {
					a.SyncStatus = entity.AccountSyncStatusDownload
				} else if a.SyncUpload {
					a.SyncStatus = entity.AccountSyncStatusUpload
				} else {
					a.SyncStatus = entity.AccountSyncStatusSynced
					a.SyncDate.Time = time.Now()
					a.SyncDate.Valid = true
				}

				event.Publish("sync.refreshed", event.Data{"account": a})
			}
		case entity.AccountSyncStatusDownload:
			if complete, err := s.download(a); err != nil {
				a.AccErrors++
				a.AccError = err.Error()
			} else if complete {
				if a.SyncUpload {
					a.SyncStatus = entity.AccountSyncStatusUpload
				} else {
					event.Publish("sync.synced", event.Data{"account": a})
					a.SyncStatus = entity.AccountSyncStatusSynced
					a.SyncDate.Time = time.Now()
					a.SyncDate.Valid = true
				}
			}
		case entity.AccountSyncStatusUpload:
			if complete, err := s.upload(a); err != nil {
				a.AccErrors++
				a.AccError = err.Error()
			} else if complete {
				event.Publish("sync.synced", event.Data{"account": a})
				a.SyncStatus = entity.AccountSyncStatusSynced
				a.SyncDate.Time = time.Now()
				a.SyncDate.Valid = true
			}
		case entity.AccountSyncStatusSynced:
			if a.SyncDate.Valid && a.SyncDate.Time.Before(time.Now().Add(time.Duration(-1*a.SyncInterval)*time.Second)) {
				a.SyncStatus = entity.AccountSyncStatusRefresh
			}
		default:
			a.SyncStatus = entity.AccountSyncStatusRefresh
		}

		if mutex.Sync.Canceled() {
			return nil
		}

		if err := db.Save(&a).Error; err != nil {
			log.Errorf("sync: %s", err.Error())
		}
	}

	return err
}
