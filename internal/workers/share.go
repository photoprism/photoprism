package workers

import (
	"fmt"
	"path"
	"runtime/debug"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/remote"
	"github.com/photoprism/photoprism/internal/remote/webdav"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
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

		if a.AccType != remote.ServiceWebDAV {
			continue
		}

		files, err := query.FileShares(a.ID, entity.FileShareNew)

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

		client, err := webdav.NewClient(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout))

		if err != nil {
			return err
		}

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

			dir := path.Dir(file.RemoteName)

			// Ensure remote folder exists.
			if err := client.MkdirAll(dir); err != nil {
				log.Debugf("share: %s", err)
			}

			srcFileName := photoprism.FileName(file.File.FileRoot, file.File.FileName)

			if fs.ImageJPEG.Equal(file.File.FileType) && size.Width > 0 && size.Height > 0 {
				srcFileName, err = thumb.FromFile(srcFileName, file.File.FileHash, w.conf.ThumbCachePath(), size.Width, size.Height, file.File.FileOrientation, size.Options...)

				if err != nil {
					w.logErr(err)
					continue
				}
			}

			if err := client.Upload(srcFileName, file.RemoteName); err != nil {
				w.logErr(err)
				file.Errors++
				file.Error = err.Error()
			} else {
				log.Infof("share: uploaded %s to %s", file.RemoteName, a.AccName)
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
	}

	// Remove previously shared files if expired
	for _, a := range accounts {
		if mutex.ShareWorker.Canceled() {
			return nil
		}

		if a.AccType != remote.ServiceWebDAV {
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

		client, err := webdav.NewClient(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout))

		if err != nil {
			return err
		}

		for _, file := range files {
			if mutex.ShareWorker.Canceled() {
				return nil
			}

			if err := client.Delete(file.RemoteName); err != nil {
				file.Errors++
				file.Error = err.Error()
			} else {
				log.Infof("share: removed %s from %s", file.RemoteName, a.AccName)
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
