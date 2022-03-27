package workers

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/remote"
	"github.com/photoprism/photoprism/internal/remote/webdav"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Updates the local list of remote files so that they can be downloaded in batches
func (worker *Sync) refresh(a entity.Account) (complete bool, err error) {
	if a.AccType != remote.ServiceWebDAV {
		return false, nil
	}

	client := webdav.New(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout))

	subDirs, err := client.Directories(a.SyncPath, true, webdav.MaxRequestDuration)

	if err != nil {
		log.Error(err)
		return false, err
	}

	dirs := append(subDirs.Abs(), a.SyncPath)

	for _, dir := range dirs {
		if mutex.SyncWorker.Canceled() {
			return false, nil
		}

		files, err := client.Files(dir)

		if err != nil {
			log.Error(err)
			return false, err
		}

		for _, file := range files {
			if mutex.SyncWorker.Canceled() {
				return false, nil
			}

			f := entity.NewFileSync(a.ID, file.Abs)

			f.Status = entity.FileSyncIgnore
			f.RemoteDate = file.Date
			f.RemoteSize = file.Size

			// Select supported types for download
			mediaType := fs.GetMediaType(file.Name)
			switch mediaType {
			case fs.MediaImage, fs.MediaSidecar:
				f.Status = entity.FileSyncNew
			case fs.MediaRaw, fs.MediaVideo:
				if a.SyncRaw {
					f.Status = entity.FileSyncNew
				}
			}

			f = entity.FirstOrCreateFileSync(f)

			if f == nil {
				log.Errorf("sync: file sync entity should not be nil - bug?")
				continue
			}

			if f.Status == entity.FileSyncIgnore && a.SyncRaw && (mediaType == fs.MediaRaw || mediaType == fs.MediaVideo) {
				worker.logError(f.Update("Status", entity.FileSyncNew))
			}

			if f.Status == entity.FileSyncDownloaded && !f.RemoteDate.Equal(file.Date) {
				worker.logError(f.Updates(map[string]interface{}{
					"Status":     entity.FileSyncNew,
					"RemoteDate": file.Date,
					"RemoteSize": file.Size,
				}))
			}
		}
	}

	return true, nil
}
