package workers

import (
	"path"
	"unicode/utf8"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/service/webdav"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// refresh updates the local queue of eligible remote files.
func (w *Sync) refresh(a entity.Service) (complete bool, err error) {
	if a.AccType != service.WebDAV {
		return false, nil
	}

	if webdav.SkipSyncPath(a.SyncPath) {
		log.Tracef("sync: skipping excluded path %s for service %s (refresh)", clean.Log(a.SyncPath), clean.Log(a.AccName))
		return true, nil
	}

	// Uploads that failed in the previous sync cycle are tried again, within the retry limit.
	w.logAccountErr(entity.RetryFileSyncUploads(a.ID, a.RetryLimit), a)

	client, err := webdav.NewClient(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout), w.conf.ServicesCIDR())

	if err != nil {
		return false, err
	}

	// Ensure remote folder exists.
	if err = client.MkdirAll(a.SyncPath); err != nil {
		log.Debugf("sync: %s (create remote folder for %s)", clean.Error(err), clean.Log(a.AccName))
	}

	subDirs, err := client.Directories(a.SyncPath, true, webdav.MaxRequestDuration)

	if err != nil {
		log.Errorf("sync: %s (list remote folders of %s)", clean.Error(err), clean.Log(a.AccName))
		return false, err
	}

	// The folder listing may include the sync folder itself, which is listed once.
	dirs := append(subDirs.Abs(), a.SyncPath)
	listed := make(map[string]bool, len(dirs))

	// Remote files whose names are too long are skipped with one warning per refresh.
	var skipped skippedNames

	defer skipped.Warn("sync: skipped %s in %s whose names are too long (%s)", "remote file", a.AccName)

	// PostgreSQL cannot store names that are not valid UTF-8, so those files are skipped with one warning as well.
	var invalid skippedNames

	defer invalid.Warn("sync: skipped %s in %s whose names are not valid UTF-8 (%s)", "remote file", a.AccName)

	validNames := entity.DbDialect() == dsn.DialectPostgreSQL

	// Queued files that this refresh does not list are removed from the queue once it completes.
	unlisted, err := queuedFileSyncNames(a.ID)

	if err != nil {
		return false, err
	}

	for _, dir := range dirs {
		if key := path.Clean("/" + dir); listed[key] {
			continue
		} else {
			listed[key] = true
		}

		if webdav.SkipSyncPath(dir) {
			log.Debugf("sync: skipping excluded path %s", clean.Log(dir))
			continue
		}
		if mutex.SyncWorker.Canceled() {
			return false, nil
		}

		files, err := client.Files(dir, false)

		if err != nil {
			log.Errorf("sync: %s (list remote files of %s)", clean.Error(err), clean.Log(a.AccName))
			return false, err
		}

		for _, file := range files {
			delete(unlisted, file.Abs)

			if webdav.SkipSyncPath(file.Abs) {
				log.Debugf("sync: skipping excluded path %s", clean.Log(file.Abs))
				continue
			}
			if mutex.SyncWorker.Canceled() {
				return false, nil
			}
			if syncNameTooLong(file.Abs) {
				skipped.Add(file.Abs)
				continue
			} else if validNames && !utf8.ValidString(file.Abs) {
				invalid.Add(file.Abs)
				continue
			}

			f := entity.NewFileSync(a.ID, file.Abs)

			f.Status = entity.FileSyncIgnore
			f.RemoteDate = file.Date
			f.RemoteSize = file.Size

			// Select supported types for download.
			content := media.FromName(file.Name)
			yamlFile := fs.FileType(file.Name) == fs.SidecarYaml
			switch content {
			case media.Image, media.Sidecar, media.Vector, media.Document, media.Live, media.Animated:
				if a.SyncYamlEnabled() || !yamlFile {
					f.Status = entity.FileSyncNew
				}
			case media.Raw, media.Video:
				if a.SyncRaw {
					f.Status = entity.FileSyncNew
				}
			}

			f = entity.FirstOrCreateFileSync(f)

			if f == nil {
				log.Errorf("sync: failed to record %s from %s", clean.Log(file.Abs), clean.Log(a.AccName))
				continue
			}

			if f.Status == entity.FileSyncIgnore && (a.SyncRaw && (content == media.Raw || content == media.Video) || a.SyncYamlEnabled() && yamlFile) {
				w.logFileErr(f.Update("Status", entity.FileSyncNew), file.Abs, a)
			}

			// Keep the listed date and size of queued files current, since a replaced file may keep its date.
			if f.Status == entity.FileSyncNew && (!f.RemoteDate.Equal(file.Date) || f.RemoteSize != file.Size) {
				w.logFileErr(f.Updates(entity.Values{
					"RemoteDate": file.Date,
					"RemoteSize": file.Size,
				}), file.Abs, a)
			}

			if f.Status == entity.FileSyncDownloaded && !f.RemoteDate.Equal(file.Date) {
				w.logFileErr(f.Updates(entity.Values{
					"Status":     entity.FileSyncNew,
					"RemoteDate": file.Date,
					"RemoteSize": file.Size,
				}), file.Abs, a)
			}
		}
	}

	// Files queued below another sync folder, or removed from the remote, are no longer downloaded.
	w.logAccountErr(dropQueuedFileSyncs(a.ID, unlisted), a)

	return true, nil
}

// queuedFileSyncNames returns the remote names of the files queued for download from an account.
func queuedFileSyncNames(serviceID uint) (map[string]struct{}, error) {
	var names []string

	if err := entity.Db().Model(&entity.FileSync{}).Where("service_id = ? AND status = ?", serviceID, entity.FileSyncNew).
		Pluck("remote_name", &names).Error; err != nil {
		return nil, err
	}

	result := make(map[string]struct{}, len(names))

	for _, name := range names {
		result[name] = struct{}{}
	}

	return result, nil
}

// dropQueuedFileSyncs removes the queued downloads of an account with the specified remote names.
func dropQueuedFileSyncs(serviceID uint, names map[string]struct{}) error {
	batch := make([]string, 0, 500)

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}

		err := entity.Db().Where("service_id = ? AND status = ? AND remote_name IN (?)", serviceID,
			entity.FileSyncNew, batch).Delete(&entity.FileSync{}).Error
		batch = batch[:0]

		return err
	}

	for name := range names {
		if batch = append(batch, name); len(batch) == cap(batch) {
			if err := flush(); err != nil {
				return err
			}
		}
	}

	return flush()
}
