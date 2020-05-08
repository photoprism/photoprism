package workers

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/remote"
	"github.com/photoprism/photoprism/internal/remote/webdav"
	"github.com/photoprism/photoprism/internal/thumb"
)

// Share represents a share worker.
type Share struct {
	conf *config.Config
}

// NewShare returns a new service share worker.
func NewShare(conf *config.Config) *Share {
	return &Share{conf: conf}
}

// Start starts the share worker.
func (s *Share) Start() (err error) {
	if err := mutex.Share.Start(); err != nil {
		event.Error(fmt.Sprintf("share: %s", err.Error()))
		return err
	}

	defer mutex.Share.Stop()

	f := form.AccountSearch{
		Share: true,
	}

	// Find accounts for which sharing is enabled
	accounts, err := query.Accounts(f)

	// Upload newly shared files
	for _, a := range accounts {
		if mutex.Share.Canceled() {
			return nil
		}

		if a.AccType != remote.ServiceWebDAV {
			continue
		}

		files, err := query.FileShares(a.ID, entity.FileShareNew)

		if err != nil {
			log.Errorf("share: %s", err.Error())
			continue
		}

		if len(files) == 0 {
			// No files to upload for this account
			continue
		}

		client := webdav.New(a.AccURL, a.AccUser, a.AccPass)
		existingDirs := make(map[string]string)

		for _, file := range files {
			if mutex.Share.Canceled() {
				return nil
			}

			dir := filepath.Dir(file.RemoteName)

			if _, ok := existingDirs[dir]; !ok {
				if err := client.CreateDir(dir); err != nil {
					log.Errorf("share: could not create folder %s", dir)
					continue
				}
			}

			srcFileName := path.Join(s.conf.OriginalsPath(), file.File.FileName)

			if a.ShareSize != "" {
				thumbType, ok := thumb.Types[a.ShareSize]

				if !ok {
					log.Errorf("share: invalid size %s", a.ShareSize)
					continue
				}

				srcFileName, err = thumb.FromFile(srcFileName, file.File.FileHash, s.conf.ThumbPath(), thumbType.Width, thumbType.Height, thumbType.Options...)

				if err != nil {
					log.Errorf("share: %s", err)
					continue
				}
			}

			if err := client.Upload(srcFileName, file.RemoteName); err != nil {
				log.Errorf("share: %s", err.Error())
				file.Errors++
				file.Error = err.Error()
			} else {
				log.Infof("share: uploaded %s to %s", file.RemoteName, a.AccName)
				file.Errors = 0
				file.Error = ""
				file.Status = entity.FileShareShared
			}

			if a.RetryLimit >= 0 && file.Errors > a.RetryLimit {
				file.Status = entity.FileShareError
			}

			if mutex.Share.Canceled() {
				return nil
			}

			if err := entity.Db().Save(&file).Error; err != nil {
				log.Errorf("share: %s", err.Error())
			}
		}
	}

	// Remove previously shared files if expired
	for _, a := range accounts {
		if mutex.Share.Canceled() {
			return nil
		}

		if a.AccType != remote.ServiceWebDAV {
			continue
		}

		files, err := query.ExpiredFileShares(a)

		if err != nil {
			log.Errorf("share: %s", err.Error())
			continue
		}

		if len(files) == 0 {
			// No files to remove for this account
			continue
		}

		client := webdav.New(a.AccURL, a.AccUser, a.AccPass)

		for _, file := range files {
			if mutex.Share.Canceled() {
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
				log.Errorf("share: %s", err.Error())
			}
		}
	}

	return err
}
