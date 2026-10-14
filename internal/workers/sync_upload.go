package workers

import (
	"errors"
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

	// Files skipped for the length of their remote name below another sync path are checked again.
	w.logErr(entity.ResetFileSyncLongNames(a.ID, a.SyncPath))

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

	// A YAML file refused with 403 disables YAML sync unless the remote also refused another file.
	var yamlRefused, otherRefused bool

	// Files whose remote names are too long are reported in one warning per run.
	var skipped skippedNames

	defer skipped.Warn("sync: skipped upload of %s to %s because their remote names are too long (%s)", "file", a.AccName)

	for _, file := range files {
		if mutex.SyncWorker.Canceled() {
			return false, nil
		}

		if webdav.SkipSyncPath(file.FileName) {
			log.Debugf("sync: skipping excluded path %s", clean.Log(file.FileName))
			w.ignoreUpload(a, path.Join(fs.PPHiddenPathname, "sync"), file.ID)
			continue
		}

		yamlFile := fs.Type(file.FileType) == fs.SidecarYaml

		if yamlFile && (!a.SyncYamlEnabled() || yamlRefused) {
			continue
		}

		fileName := photoprism.FileName(file.FileRoot, file.FileName)
		remoteName := path.Join(a.SyncPath, file.FileName)
		remoteDir := path.Dir(remoteName)

		// A remote name that is too long is recorded for the sync path, so the file is not selected again.
		if entity.RemoteNameTooLong(remoteName) {
			skipped.Add(file.FileName)
			w.ignoreUpload(a, entity.FileSyncLongNamePrefix(a.SyncPath), file.ID)
			continue
		}

		// Ensure remote folder exists.
		if err = client.MkdirAll(remoteDir); err != nil {
			log.Debugf("sync: %s", err)
		}

		if err = client.Upload(fileName, remoteName); errors.Is(err, webdav.ErrForbidden) {
			yamlRefused, otherRefused = yamlRefused || yamlFile, otherRefused || !yamlFile
			w.logErr(err)
			continue
		} else if err != nil {
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

	if yamlRefused && !otherRefused {
		log.Warnf("sync: disabled YAML sidecar files for %s because the remote server refused to store them", clean.Log(a.AccName))
		w.logErr(a.Update("SyncYaml", -1))
	}

	return false, nil
}

// ignoreUpload records a local file as ignored for an account under a placeholder name in dir, so it is not
// selected for upload again.
func (w *Sync) ignoreUpload(a entity.Service, dir string, fileID uint) {
	ignored := entity.NewFileSync(a.ID, path.Join(dir, strconv.FormatUint(uint64(fileID), 10)))
	ignored.FileID, ignored.Status = fileID, entity.FileSyncIgnore
	w.logErr(entity.Db().Save(ignored).Error)
}
