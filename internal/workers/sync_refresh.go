package workers

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/remote"
	"github.com/photoprism/photoprism/internal/remote/webdav"
	"github.com/photoprism/photoprism/pkg/media"
)

// Updates the local list of remote files so that they can be downloaded in batches
func (w *Sync) refresh(a entity.Service) (complete bool, err error) {
	if a.AccType != remote.ServiceWebDAV {
		return false, nil
	}

	client, err := webdav.NewClient(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout))

	if err != nil {
		return false, err
	}

	// Ensure remote folder exists.
	if err = client.MkdirAll(a.SyncPath); err != nil {
		log.Debugf("sync: %s", err)
	}

	subDirs, err := client.Directories(a.SyncPath, true, webdav.MaxRequestDuration)

	if err != nil {
		log.Errorf("sync: %s", err)
		return false, err
	}

	dirs := append(subDirs.Abs(), a.SyncPath)

	for _, dir := range dirs {
		if mutex.SyncWorker.Canceled() {
			return false, nil
		}

		files, err := client.Files(dir, false)

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
			content := media.FromName(file.Name)
			switch content {
			case media.Image, media.Sidecar:
				f.Status = entity.FileSyncNew
			case media.Raw, media.Video:
				if a.SyncRaw {
					f.Status = entity.FileSyncNew
				}
			}

			f = entity.FirstOrCreateFileSync(f)

			if f == nil {
				log.Errorf("sync: file sync entity should not be nil - possible bug")
				continue
			}

			if f.Status == entity.FileSyncIgnore && a.SyncRaw && (content == media.Raw || content == media.Video) {
				w.logError(f.Update("Status", entity.FileSyncNew))
			}

			if f.Status == entity.FileSyncDownloaded && !f.RemoteDate.Equal(file.Date) {
				w.logError(f.Updates(map[string]interface{}{
					"Status":     entity.FileSyncNew,
					"RemoteDate": file.Date,
					"RemoteSize": file.Size,
				}))
			}
		}
	}

	return true, nil
}
