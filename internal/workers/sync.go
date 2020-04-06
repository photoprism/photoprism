package workers

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/remote"
	"github.com/photoprism/photoprism/internal/remote/webdav"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Sync represents a sync worker.
type Sync struct {
	conf *config.Config
	q    *query.Query
}

type Downloads map[string][]entity.FileSync

// NewSync returns a new service sync worker.
func NewSync(conf *config.Config) *Sync {
	return &Sync{
		conf: conf,
		q:    query.New(conf.Db()),
	}
}

// DownloadPath returns a temporary download path.
func (s *Sync) DownloadPath() string {
	return s.conf.TempPath() + "/sync"
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

func (s *Sync) relatedDownloads(a entity.Account) (result Downloads, err error) {
	result = make(Downloads)
	maxResults := 1000

	// Get remote files from database
	files, err := s.q.FileSyncs(a.ID, entity.FileSyncNew, maxResults)

	if err != nil {
		return result, err
	}

	// Group results by directory and base name
	for i, file := range files {
		k := fs.AbsBase(file.RemoteName)

		result[k] = append(result[k], file)

		// Skip last 50 to make sure we see all related files
		if i > (maxResults - 50) {
			return result, nil
		}
	}

	return result, nil
}

func (s *Sync) download(a entity.Account) (complete bool, err error) {
	db := s.conf.Db()

	// Set up index worker
	indexJobs := make(chan photoprism.IndexJob)
	go photoprism.IndexWorker(indexJobs)
	defer close(indexJobs)

	// Set up import worker
	importJobs := make(chan photoprism.ImportJob)
	go photoprism.ImportWorker(importJobs)
	defer close(importJobs)

	relatedFiles, err := s.relatedDownloads(a)

	if err != nil {
		log.Errorf("sync: %s", err.Error())
		return false, err
	}

	if len(relatedFiles) == 0 {
		log.Infof("sync: download complete for %s", a.AccName)
		event.Publish("sync.downloaded", event.Data{"account": a})
		return true, nil
	}

	log.Infof("sync: downloading from %s", a.AccName)

	client := webdav.New(a.AccURL, a.AccUser, a.AccPass)

	var baseDir string

	if a.SyncFilenames {
		baseDir = s.conf.OriginalsPath()
	} else {
		baseDir = fmt.Sprintf("%s/%d", s.DownloadPath(), a.ID)
	}

	done := make(map[string]bool)

	for _, files := range relatedFiles {
		for _, file := range files {
			if mutex.Sync.Canceled() {
				return false, nil
			}

			if file.Errors > a.RetryLimit {
				log.Warnf("sync: downloading %s failed more than %d times", file.RemoteName, a.RetryLimit)
				continue
			}

			localName := baseDir + file.RemoteName

			if err := client.Download(file.RemoteName, localName, false); err != nil {
				log.Errorf("sync: %s", err.Error())
				file.Errors++
				file.Error = err.Error()
			} else {
				log.Infof("sync: downloaded %s from %s", file.RemoteName, a.AccName)
				file.Status = entity.FileSyncDownloaded
			}

			if mutex.Sync.Canceled() {
				return false, nil
			}

			if err := db.Save(&file).Error; err != nil {
				log.Errorf("sync: %s", err.Error())
			}
		}

		for _, file := range files {
			mf, err := photoprism.NewMediaFile(baseDir + file.RemoteName)

			if err != nil || !mf.IsPhoto() {
				continue
			}

			related, err := mf.RelatedFiles()

			if err != nil {
				log.Warnf("sync: %s", err.Error())
				continue
			}

			var rf photoprism.MediaFiles

			for _, f := range related.Files {
				if done[f.FileName()] {
					continue
				}

				rf = append(rf, f)
				done[f.FileName()] = true
			}

			done[mf.FileName()] = true
			related.Files = rf

			if a.SyncFilenames {
				log.Infof("sync: indexing %s and related files", file.RemoteName)
				indexJobs <- photoprism.IndexJob{
					FileName: mf.FileName(),
					Related:  related,
					IndexOpt: photoprism.IndexOptionsAll(),
					Ind:      service.Index(),
				}
			} else {
				log.Infof("sync: importing %s and related files", file.RemoteName)
				importJobs <- photoprism.ImportJob{
					FileName:  mf.FileName(),
					Related:   related,
					IndexOpt:  photoprism.IndexOptionsAll(),
					ImportOpt: photoprism.ImportOptionsMove(baseDir),
					Imp:       service.Import(),
				}
			}
		}
	}

	return false, nil
}

func (s *Sync) upload(a entity.Account) (complete bool, err error) {
	// TODO: Not implemented yet
	return false, nil
}
