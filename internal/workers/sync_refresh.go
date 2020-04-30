package workers

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/remote"
	"github.com/photoprism/photoprism/internal/remote/webdav"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Updates the local list of remote files so that they can be downloaded in batches
func (s *Sync) refresh(a entity.Account) (complete bool, err error) {
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

			f.Status = entity.FileSyncIgnore
			f.RemoteDate = file.Date
			f.RemoteSize = file.Size

			// Select supported types for download
			mediaType := fs.GetMediaType(file.Name)
			switch mediaType {
			case fs.MediaImage:
				f.Status = entity.FileSyncNew
			case fs.MediaSidecar:
				f.Status = entity.FileSyncNew
			case fs.MediaRaw:
				if a.SyncRaw {
					f.Status = entity.FileSyncNew
				}
			}

			f.FirstOrCreate()

			if f.Status == entity.FileSyncIgnore && mediaType == fs.MediaRaw && a.SyncRaw {
				f.Status = entity.FileSyncNew
				db.Save(&f)
			}

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
