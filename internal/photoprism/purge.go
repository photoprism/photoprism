package photoprism

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Purge represents a worker that removes missing files from search results.
type Purge struct {
	conf  *config.Config
	files *Files
}

// NewPurge returns a new purge worker.
func NewPurge(conf *config.Config, files *Files) *Purge {
	instance := &Purge{
		conf:  conf,
		files: files,
	}

	return instance
}

// Start removes missing files from search results.
func (w *Purge) Start(opt PurgeOptions) (purgedFiles map[string]bool, purgedPhotos map[string]bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("purge: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	var ignore fs.Done

	if opt.Ignore != nil {
		ignore = opt.Ignore
	} else {
		ignore = make(fs.Done)
	}

	purgedFiles = make(map[string]bool)
	purgedPhotos = make(map[string]bool)

	if err := mutex.MainWorker.Start(); err != nil {
		log.Warnf("purge: %s (start)", err.Error())
		return purgedFiles, purgedPhotos, err
	}

	defer mutex.MainWorker.Stop()

	limit := 500
	offset := 0

	for {
		files, err := query.Files(limit, offset, opt.Path, true)

		if err != nil {
			return purgedFiles, purgedPhotos, err
		}

		if len(files) == 0 {
			break
		}

		for _, file := range files {
			if mutex.MainWorker.Canceled() {
				return purgedFiles, purgedPhotos, errors.New("purge canceled")
			}

			fileName := FileName(file.FileRoot, file.FileName)

			if ignore[fileName].Exists() || purgedFiles[fileName] {
				continue
			}

			if file.FileMissing {
				if fs.FileExists(fileName) {
					if opt.Dry {
						log.Infof("purge: found %s", txt.Quote(file.FileName))
						continue
					}

					if err := file.Found(); err != nil {
						log.Errorf("purge: %s", err)
					} else {
						log.Infof("purge: found %s", txt.Quote(file.FileName))
					}
				}
			} else if !fs.FileExists(fileName) {
				if opt.Dry {
					purgedFiles[fileName] = true
					log.Infof("purge: file %s would be flagged as missing", txt.Quote(file.FileName))
					continue
				}

				wasPrimary := file.FilePrimary

				if err := file.Purge(); err != nil {
					log.Errorf("purge: %s", err)
					continue
				}

				w.files.Remove(file.FileName, file.FileRoot)
				purgedFiles[fileName] = true
				log.Infof("purge: flagged file %s as missing", txt.Quote(file.FileName))

				if !wasPrimary {
					continue
				}

				if err := query.SetPhotoPrimary(file.PhotoUID, ""); err != nil {
					log.Infof("purge: %s", err)
				}
			}
		}

		if mutex.MainWorker.Canceled() {
			return purgedFiles, purgedPhotos, errors.New("purge canceled")
		}

		offset += limit

		time.Sleep(50 * time.Millisecond)
	}

	limit = 500
	offset = 0

	for {
		files, err := query.Duplicates(limit, offset, opt.Path)

		if err != nil {
			return purgedFiles, purgedPhotos, err
		}

		if len(files) == 0 {
			break
		}

		for _, file := range files {
			if mutex.MainWorker.Canceled() {
				return purgedFiles, purgedPhotos, errors.New("purge canceled")
			}

			fileName := FileName(file.FileRoot, file.FileName)

			if ignore[fileName].Exists() || purgedFiles[fileName] {
				continue
			}

			if !fs.FileExists(fileName) {
				if opt.Dry {
					purgedFiles[fileName] = true
					log.Infof("purge: duplicate %s would be removed", txt.Quote(file.FileName))
					continue
				}

				if err := file.Purge(); err != nil {
					log.Errorf("purge: %s", err)
				} else {
					w.files.Remove(file.FileName, file.FileRoot)
					purgedFiles[fileName] = true
					log.Infof("purge: removed duplicate %s", txt.Quote(file.FileName))
				}
			}
		}

		if mutex.MainWorker.Canceled() {
			return purgedFiles, purgedPhotos, errors.New("purge canceled")
		}

		offset += limit

		time.Sleep(50 * time.Millisecond)
	}

	limit = 500
	offset = 0

	for {
		photos, err := query.PhotosMissing(limit, offset)

		if err != nil {
			return purgedFiles, purgedPhotos, err
		}

		if len(photos) == 0 {
			break
		}

		for _, photo := range photos {
			if mutex.MainWorker.Canceled() {
				return purgedFiles, purgedPhotos, errors.New("purge canceled")
			}

			if purgedPhotos[photo.PhotoUID] {
				continue
			}

			if opt.Dry {
				purgedPhotos[photo.PhotoUID] = true
				log.Infof("purge: %s would be removed", txt.Quote(photo.PhotoName))
				continue
			}

			if err := photo.Delete(opt.Hard); err != nil {
				log.Errorf("purge: %s", err)
			} else {
				purgedPhotos[photo.PhotoUID] = true

				if opt.Hard {
					log.Infof("purge: permanently deleted %s", txt.Quote(photo.PhotoName))
				} else {
					log.Infof("purge: flagged photo %s as deleted", txt.Quote(photo.PhotoName))
				}

				// Remove files from lookup table.
				for _, file := range photo.AllFiles() {
					w.files.Remove(file.FileName, file.FileRoot)
				}
			}
		}

		if mutex.MainWorker.Canceled() {
			return purgedFiles, purgedPhotos, errors.New("purge canceled")
		}

		offset += limit

		time.Sleep(50 * time.Millisecond)
	}

	log.Info("purge: searching index for unassigned primary files")

	if err := query.FixPrimaries(); err != nil {
		log.Errorf("purge: %s (fix primary files)", err.Error())
	}

	log.Info("purge: searching index for hidden media files")

	if err := query.ResetPhotoQuality(); err != nil {
		return purgedFiles, purgedPhotos, err
	}

	if err := query.PurgeOrphans(); err != nil {
		log.Errorf("purge: %s (orphans)", err)
	}

	if err := query.UpdateMissingAlbumEntries(); err != nil {
		log.Errorf("purge: %s (album entries)", err)
	}

	if err := entity.UpdatePhotoCounts(); err != nil {
		log.Errorf("purge: %s (photo counts)", err)
	}

	return purgedFiles, purgedPhotos, nil
}

// Cancel stops the current operation.
func (w *Purge) Cancel() {
	mutex.MainWorker.Cancel()
}
