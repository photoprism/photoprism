package workers

import (
	"errors"
	"fmt"
	"path"
	"runtime/debug"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/service/webdav"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// Share represents a share worker.
type Share struct {
	conf *config.Config
}

// NewShare returns a new share worker.
func NewShare(conf *config.Config) *Share {
	return &Share{conf: conf}
}

// logErr logs an error message if err is not nil.
func (w *Share) logErr(err error) {
	if err != nil {
		log.Errorf("share: %s", err.Error())
	}
}

// Start starts the share worker.
func (w *Share) Start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("share: %s (worker panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err := mutex.ShareWorker.Start(); err != nil {
		return err
	}

	defer mutex.ShareWorker.Stop()

	f := form.SearchServices{
		Share: true,
	}

	// Find accounts for which sharing is enabled
	accounts, err := search.Accounts(f)

	// Upload newly shared files
	for _, a := range accounts {
		if mutex.ShareWorker.Canceled() {
			return nil
		}

		if a.AccType != service.WebDAV {
			continue
		}

		files, err := query.QueuedFileShares(a)

		if err != nil {
			w.logErr(err)
			continue
		}

		if len(files) == 0 {
			// No files to upload for this account
			continue
		}

		size := thumb.Size{}

		if a.ShareSize != "" {
			if s, ok := thumb.Sizes[thumb.Name(a.ShareSize)]; ok {
				size = s
			} else {
				size = thumb.SizeFit1920
			}
		}

		client, err := webdav.NewClient(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout), w.conf.ServicesCIDR())

		if err != nil {
			return err
		}

		// Count failed transfers so a single UI notification can report them per service,
		// since the manual upload request returns before the worker runs (#5738).
		var uploadErrors int

		// A YAML file refused with 403 disables YAML sync unless the remote also refused another file.
		var refusedYaml []entity.FileShare
		var otherRefused bool

		for _, file := range files {
			if mutex.ShareWorker.Canceled() {
				return nil
			}

			// Skip deleted files.
			if file.File == nil || file.FileID <= 0 {
				log.Warnf("share: %s cannot be uploaded because it has been deleted", clean.Log(file.RemoteName))
				file.Status = entity.FileShareError
				file.Error = "file not found"
				file.Errors++
				w.logErr(entity.Db().Save(&file).Error)
				continue
			}

			if webdav.SkipSyncPath(file.File.FileName) || webdav.SkipSyncPath(file.RemoteName) {
				log.Debugf("share: skipping excluded transfer %s to %s", clean.Log(file.File.FileName), clean.Log(file.RemoteName))
				file.Status, file.Error, file.Errors = entity.FileShareIgnore, "", 0
				w.logErr(entity.Db().Save(&file).Error)
				continue
			}

			yamlFile := fs.SidecarYaml.Equal(file.File.FileType)

			// Further YAML files stay queued once the remote server refused one.
			if yamlFile && len(refusedYaml) > 0 {
				continue
			}

			dir := path.Dir(file.RemoteName)

			// Ensure remote folder exists.
			if err := client.MkdirAll(dir); err != nil {
				log.Debugf("share: %s", err)
			}

			srcFileName := photoprism.FileName(file.File.FileRoot, file.File.FileName)

			if fs.ImageJpeg.Equal(file.File.FileType) && size.Width > 0 && size.Height > 0 {
				srcFileName, err = thumb.FromFile(srcFileName, file.File.FileHash, w.conf.ThumbCachePath(), size.Width, size.Height, file.File.FileOrientation, size.Options...)

				if err != nil {
					w.logErr(err)
					continue
				}
			}

			if err = client.Upload(srcFileName, file.RemoteName); yamlFile && errors.Is(err, webdav.ErrForbidden) {
				w.logErr(err)
				file.Error = err.Error()
				refusedYaml = append(refusedYaml, file)
				continue
			} else if err != nil {
				w.logErr(err)
				otherRefused = otherRefused || errors.Is(err, webdav.ErrForbidden)
				uploadErrors++
				file.Errors++
				file.Error = err.Error()
			} else {
				log.Infof("share: uploaded %s to %s", clean.Log(file.RemoteName), clean.Log(a.AccName))
				file.Errors = 0
				file.Error = ""
				file.Status = entity.FileShareShared
			}

			// Failed too often?
			if a.RetryLimit > 0 && file.Errors > a.RetryLimit {
				file.Status = entity.FileShareError
			}

			if mutex.ShareWorker.Canceled() {
				return nil
			}

			w.logErr(entity.Db().Save(&file).Error)
		}

		if len(refusedYaml) > 0 && !otherRefused {
			log.Warnf("share: disabled YAML sidecar files for %s because the remote server refused to store them", clean.Log(a.AccName))
			w.logErr(a.Update("SyncYaml", -1))
		} else {
			for _, file := range refusedYaml {
				uploadErrors++
				file.Errors++

				if a.RetryLimit > 0 && file.Errors > a.RetryLimit {
					file.Status = entity.FileShareError
				}

				w.logErr(entity.Db().Save(&file).Error)
			}
		}

		// Notify the user if any transfer to this service failed, since the manual upload
		// request already returned before the worker ran.
		if uploadErrors > 0 {
			event.ErrorMsg(i18n.ErrUploadToServiceFailed, a.AccName)
		}
	}

	// Remove previously shared files if expired
	for _, a := range accounts {
		if mutex.ShareWorker.Canceled() {
			return nil
		}

		if a.AccType != service.WebDAV {
			continue
		}

		files, err := query.ExpiredFileShares(a)

		if err != nil {
			w.logErr(err)
			continue
		}

		if len(files) == 0 {
			// No files to remove for this account
			continue
		}

		client, err := webdav.NewClient(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout), w.conf.ServicesCIDR())

		if err != nil {
			return err
		}

		for _, file := range files {
			if mutex.ShareWorker.Canceled() {
				return nil
			}

			if webdav.SkipSyncPath(file.RemoteName) || webdav.UnsafeSyncPath(file.RemoteName) {
				file.Status = entity.FileShareError
				file.Error = "remote copy retained: removal blocked by path policy"
				file.Errors++
				log.Warnf("share: expired remote copy %s on service %s retained by path policy; manual removal required", clean.Log(file.RemoteName), clean.Log(a.AccName))
				w.logErr(entity.Db().Save(&file).Error)
				continue
			}

			if err := client.Delete(file.RemoteName); err != nil {
				file.Errors++
				file.Error = err.Error()
			} else {
				log.Infof("share: removed %s from %s", clean.Log(file.RemoteName), clean.Log(a.AccName))
				file.Errors = 0
				file.Error = ""
				file.Status = entity.FileShareRemoved
			}

			if err := entity.Db().Save(&file).Error; err != nil {
				w.logErr(err)
			}
		}
	}

	return err
}
