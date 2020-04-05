package photoprism

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
	"github.com/photoprism/photoprism/internal/remote/webdav"
)

// Sync represents a sync worker.
type Sync struct {
	conf *config.Config
}

// NewSync returns a new service sync worker.
func NewSync(conf *config.Config) *Sync {
	return &Sync{conf: conf}
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
	q := query.New(db)

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
			if complete, err := s.getRemoteFiles(a); err != nil {
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
			}
		case entity.AccountSyncStatusDownload:
			if complete, err := s.download(a); err != nil {
				a.AccErrors++
				a.AccError = err.Error()
			} else if complete && a.SyncUpload {
				a.SyncStatus = entity.AccountSyncStatusUpload
			} else if complete {
				a.SyncStatus = entity.AccountSyncStatusSynced
				a.SyncDate.Time = time.Now()
				a.SyncDate.Valid = true
			}
		case entity.AccountSyncStatusUpload:
			if complete, err := s.upload(a); err != nil {
				a.AccErrors++
				a.AccError = err.Error()
			} else if complete {
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

func (s *Sync) getRemoteFiles(a entity.Account) (complete bool, err error) {
	if a.AccType != remote.ServiceWebDAV {
		return false, nil
	}

	db := s.conf.Db()
	client := webdav.New(a.AccURL, a.AccUser, a.AccPass)

	subDirs, err := client.Directories(a.SyncPath, true)

	if err != nil {
		log.Error(err)
		return false, err
	}

	dirs := append(subDirs.Abs(), a.SyncPath)

	for _, dir := range dirs {
		if mutex.Sync.Canceled() {
			return false, nil
		}

		files, err := client.Files(dir)

		if err != nil {
			log.Error(err)
			return false, err
		}

		for _, file := range files {
			if mutex.Sync.Canceled() {
				return false, nil
			}

			f := entity.NewFileSync(a.ID, file.Abs)
			f.RemoteDate = file.Date
			f.RemoteSize = file.Size
			f.FirstOrCreate(db)

			if f.Status == entity.FileSyncDownloaded && !f.RemoteDate.Equal(file.Date) {
				f.Status = entity.FileSyncNew
				f.RemoteDate = file.Date
				f.RemoteSize = file.Size
				db.Save(&f)
			}
		}
	}

	return true, nil
}

func (s *Sync) download(a entity.Account) (complete bool, err error) {
	db := s.conf.Db()
	q := query.New(db)

	files, err := q.FileSyncs(a.ID, entity.FileSyncNew)

	if err != nil {
		log.Errorf("sync: %s", err.Error())
		return false, err
	}

	if len(files) == 0 {
		// TODO: Subscribe event to start indexing / importing
		event.Publish("sync.downloaded", event.Data{"account": a})
		return true, nil
	}

	client := webdav.New(a.AccURL, a.AccUser, a.AccPass)

	var baseDir string

	if a.SyncFilenames {
		baseDir = s.conf.OriginalsPath()
	} else {
		baseDir = fmt.Sprintf("%s/sync/%d", s.conf.ImportPath(), a.ID)
	}

	for _, file := range files {
		if mutex.Sync.Canceled() {
			return false, nil
		}

		if file.Errors > a.RetryLimit {
			log.Warnf("sync: downloading %s failed more than %d times", file.RemoteName, a.RetryLimit)
			continue
		}

		localName := baseDir + file.RemoteName

		if err := client.Download(file.RemoteName, localName); err != nil {
			file.Errors++
			file.Error = err.Error()
		} else {
			file.Status = entity.FileSyncDownloaded
		}

		if mutex.Sync.Canceled() {
			return false, nil
		}

		if err := db.Save(&file).Error; err != nil {
			log.Errorf("sync: %s", err.Error())
		}
	}

	return false, nil
}

func (s *Sync) upload(a entity.Account) (complete bool, err error) {
	// TODO: Not implemented yet
	return false, nil
}
