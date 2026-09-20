package workers

import (
	"path"
	"strconv"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service/webdav"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// upload transfers eligible local files to a remote account.
func (w *Sync) upload(a entity.Service) (complete bool, err error) {
	if webdav.SkipSyncPath(a.SyncPath) {
		log.Tracef("sync: skipping excluded path %s for service %s (upload)", clean.Log(a.SyncPath), clean.Log(a.AccName))
		return true, nil
	}

	maxResults := 250

	// Get upload file list from database
	files, err := query.AccountUploads(a, maxResults)

	if err != nil {
		return false, err
	}

	if len(files) == 0 {
		log.Infof("sync: upload complete for %s", a.AccName)
		event.Publish("sync.uploaded", event.Data{"account": a})
		return true, nil
	}

	client, err := webdav.NewClient(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout), w.conf.ServicesCIDR())

	if err != nil {
		return false, err
	}

	for _, file := range files {
		if mutex.SyncWorker.Canceled() {
			return false, nil
		}

		if webdav.SkipSyncPath(file.FileName) {
			log.Debugf("sync: skipping excluded path %s", clean.Log(file.FileName))
			ignored := entity.NewFileSync(a.ID, path.Join(fs.PPHiddenPathname, "sync", strconv.FormatUint(uint64(file.ID), 10)))
			ignored.FileID, ignored.Status = file.ID, entity.FileSyncIgnore
			w.logErr(entity.Db().Save(ignored).Error)
			continue
		}

		fileName := photoprism.FileName(file.FileRoot, file.FileName)
		remoteName := path.Join(a.SyncPath, file.FileName)
		remoteDir := path.Dir(remoteName)

		// Ensure remote folder exists.
		if err = client.MkdirAll(remoteDir); err != nil {
			log.Debugf("sync: %s", err)
		}

		if err = client.Upload(fileName, remoteName); err != nil {
			w.logErr(err)
			continue // try again next time
		}

		log.Infof("sync: uploaded %s to %s (%s)", clean.Log(file.FileName), clean.Log(remoteName), a.AccName)

		fileSync := entity.NewFileSync(a.ID, remoteName)
		fileSync.Status = entity.FileSyncUploaded
		fileSync.RemoteDate = time.Now()
		fileSync.RemoteSize = file.FileSize
		fileSync.FileID = file.ID
		fileSync.Error = ""
		fileSync.Errors = 0

		if mutex.SyncWorker.Canceled() {
			return false, nil
		}

		w.logErr(entity.Db().Save(&fileSync).Error)
	}

	return false, nil
}
