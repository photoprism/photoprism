package workers

import (
	"path"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/service/webdav"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// refresh updates the local queue of eligible remote files.
func (w *Sync) refresh(a entity.Service) (complete bool, err error) {
	if a.AccType != service.WebDAV {
		return false, nil
	}

	if webdav.SkipSyncPath(a.SyncPath) {
		log.Warnf("sync: skipped %s because its sync folder %s is excluded", clean.Log(a.AccName), clean.Log(a.SyncPath))
		return true, syncNotice("sync folder is not supported")
	}

	// Uploads that failed in the previous sync cycle are tried again, within the retry limit.
	w.logAccountErr(entity.RetryFileSyncUploads(a.ID, a.RetryLimit), a)

	// Files that keep their remote paths in originals are not downloaded from a sync folder whose name is not
	// portable, and files queued before are removed from the queue, since nothing is listed.
	if a.SyncFilenames && !portableSyncName("/", a.SyncPath) {
		if a.SyncDownload {
			log.Warnf("sync: skipped downloads from %s because sync folder %s has unsupported characters",
				clean.Log(a.AccName), clean.Log(a.SyncPath))
		}

		queued, queuedErr := queuedFileSyncNames(a.ID)

		if queuedErr != nil {
			return false, queuedErr
		}

		if err = dropQueuedFileSyncs(a.ID, queued); err != nil {
			return false, err
		} else if a.SyncDownload {
			return true, syncNotice("sync folder is not supported for downloads")
		}

		return true, nil
	}

	// Downloads that are imported need a destination that gives names the index accepts.
	if a.SyncDownload && !a.SyncFilenames {
		if destErr := get.Import().CheckDest(w.conf.ImportDest()); destErr != nil {
			event.SystemWarn([]string{"sync", "%s"}, destErr.Error())
			log.Warnf("sync: skipped downloads from %s because the import destination is not supported", clean.Log(a.AccName))
			return true, syncNotice("import destination is not supported")
		}
	}

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

	// Remote files and folders whose names are not portable are skipped with one warning as well.
	var unsupported skippedNames

	defer unsupported.Warn("sync: skipped %s in %s with unsupported characters (%s)", "remote name", a.AccName)

	nameRoot := syncNameRoot(a)
	syncDir := path.Clean("/" + a.SyncPath)

	// Queued files that this refresh does not list are removed from the queue once it completes.
	unlisted, err := queuedFileSyncNames(a.ID)

	if err != nil {
		return false, err
	}

	for _, dir := range dirs {
		key := path.Clean("/" + dir)

		if listed[key] {
			continue
		}

		listed[key] = true

		if webdav.SkipSyncPath(dir) {
			log.Debugf("sync: skipping excluded path %s", clean.Log(dir))
			continue
		} else if key != syncDir && !portableSyncName(nameRoot, dir) {
			// The sync folder itself is listed, which matters when its files are imported.
			unsupported.Add(dir)
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
			excluded := webdav.SkipSyncPath(file.Abs)

			// A file whose name is not portable is not queued, and is removed from the queue like an unlisted one.
			if !excluded && !portableSyncName(nameRoot, file.Abs) {
				unsupported.Add(file.Abs)
				continue
			}

			delete(unlisted, file.Abs)

			if excluded {
				log.Debugf("sync: skipping excluded path %s", clean.Log(file.Abs))
				continue
			}
			if mutex.SyncWorker.Canceled() {
				return false, nil
			}
			if syncNameTooLong(file.Abs) {
				skipped.Add(file.Abs)
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

// syncNameRoot returns the remote folder whose own names are not checked: the sync folder, or the root when
// downloaded files keep their remote paths in originals.
func syncNameRoot(a entity.Service) string {
	if a.SyncFilenames {
		return "/"
	}

	return a.SyncPath
}

// portableSyncName reports whether the names in a remote path below root are portable. A path that equals
// root is checked in full.
func portableSyncName(root, name string) bool {
	root = path.Clean("/" + root)
	rel := path.Clean("/" + name)

	if strings.HasPrefix(rel, root+"/") {
		rel = rel[len(root):]
	}

	return clean.UnportablePath(strings.Trim(rel, "/")) == ""
}
