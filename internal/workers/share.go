package workers

import (
	"fmt"
	"path/filepath"
	"runtime/debug"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/remote"
	"github.com/photoprism/photoprism/internal/remote/webdav"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/internal/thumb"
)

// Share represents a share worker.
type Share struct {
	conf *config.Config
}

// NewShare returns a new share worker.
func NewShare(conf *config.Config) *Share {
	return &Share{conf: conf}
}

// logError logs an error message if err is not nil.
func (worker *Share) logError(err error) {
	if err != nil {
		log.Errorf("share: %s", err.Error())
	}
}

// Start starts the share worker.
func (worker *Share) Start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("share: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err := mutex.ShareWorker.Start(); err != nil {
		return err
	}

	defer mutex.ShareWorker.Stop()

	f := form.SearchAccounts{
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
			worker.logError(err)
			continue
		}

		if len(files) == 0 {
			// No files to upload for this account
			continue
		}

		client := webdav.New(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout))
		existingDirs := make(map[string]string)

		for _, file := range files {
			if mutex.ShareWorker.Canceled() {
				return nil
			}

			dir := filepath.Dir(file.RemoteName)

			if _, ok := existingDirs[dir]; !ok {
				if err := client.CreateDir(dir); err != nil {
					log.Errorf("share: failed creating folder %s", dir)
					continue
				}
			}

			srcFileName := photoprism.FileName(file.File.FileRoot, file.File.FileName)

			if a.ShareSize != "" {
				size, ok := thumb.Sizes[thumb.Name(a.ShareSize)]

				if !ok {
					log.Errorf("share: invalid size %s", a.ShareSize)
					continue
				}

				srcFileName, err = thumb.FromFile(srcFileName, file.File.FileHash, worker.conf.ThumbPath(), size.Width, size.Height, file.File.FileOrientation, size.Options...)

				if err != nil {
					worker.logError(err)
					continue
				}
			}

			if err := client.Upload(srcFileName, file.RemoteName); err != nil {
				worker.logError(err)
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

			worker.logError(entity.Db().Save(&file).Error)
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
			worker.logError(err)
			continue
		}

		if len(files) == 0 {
			// No files to remove for this account
			continue
		}

		client := webdav.New(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout))

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
				worker.logError(err)
			}
		}
	}

	return err
}
