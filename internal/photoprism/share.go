package photoprism

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/service/webdav"
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

	db := s.conf.Db()
	q := query.New(db)

	accounts, err := q.Accounts(f)

	for _, a := range accounts {
		if a.AccType != service.TypeWebDAV {
			continue
		}

		files, err := q.FileShares(a.ID, entity.FileShareNew)

		if err != nil {
			log.Errorf("share: %s", err.Error())
			continue
		}

		if len(files) == 0 {
			// No files to upload
			continue
		}

		client := webdav.New(a.AccURL, a.AccUser, a.AccPass)
		existingDirs := make(map[string]string)

		for _, file := range files {
			dir := filepath.Dir(file.RemoteName)

			if _, ok := existingDirs[dir]; ok == false && dir != "/" && dir != "." {
				if err := client.CreateDir(dir); err != nil {
					log.Errorf("share: could not create directory %s", dir)
					continue
				}
			}

			srcFileName := s.conf.OriginalsPath() + string(os.PathSeparator) + file.File.FileName

			if a.ShareSize != "" {
				thumbType, ok := thumb.Types[a.ShareSize]

				if !ok {
					log.Errorf("share: invalid size %s", a.ShareSize)
					continue
				}

				srcFileName, err = thumb.FromFile(srcFileName, file.File.FileHash, s.conf.ThumbnailsPath(), thumbType.Width, thumbType.Height, thumbType.Options...)

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
				file.Errors = 0
				file.Error = ""
				file.Status = entity.FileShareShared
			}

			if a.RetryLimit >= 0 && file.Errors > a.RetryLimit {
				file.Status = entity.FileShareError
			}

			if err := db.Save(&file).Error; err != nil {
				log.Errorf("share: %s", err.Error())
			}
		}
	}

	for _, a := range accounts {
		if a.AccType != service.TypeWebDAV {
			continue
		}

		files, err := q.ExpiredFileShares(a)

		if err != nil {
			log.Errorf("share: %s", err.Error())
			continue
		}

		if len(files) == 0 {
			// No files to remove
			continue
		}

		client := webdav.New(a.AccURL, a.AccUser, a.AccPass)

		for _, file := range files {
			if err := client.Delete(file.RemoteName); err != nil {
				file.Errors++
				file.Error = err.Error()
			} else {
				file.Errors = 0
				file.Error = ""
				file.Status = entity.FileShareRemoved
			}

			if err := db.Save(&file).Error; err != nil {
				log.Errorf("share: %s", err.Error())
			}
		}
	}

	return err
}
