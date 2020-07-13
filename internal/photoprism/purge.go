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
	conf *config.Config
}

// NewPurge returns a new purge worker.
func NewPurge(conf *config.Config) *Purge {
	instance := &Purge{
		conf: conf,
	}

	return instance
}

// Start removes missing files from search results.
func (prg *Purge) Start(opt PurgeOptions) (purgedFiles map[string]bool, purgedPhotos map[string]bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("purge: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	var ignore map[string]bool

	if opt.Ignore != nil {
		ignore = opt.Ignore
	} else {
		ignore = make(map[string]bool)
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

			if ignore[fileName] || purgedFiles[fileName] {
				continue
			}

			if file.FileMissing {
				if fs.FileExists(fileName) {
					if opt.Dry {
						log.Infof("purge: found %s", txt.Quote(file.FileName))
						continue
					}

					if err := file.Update("FileMissing", false); err != nil {
						log.Errorf("purge: %s", err)
					} else {
						log.Infof("purge: found %s", txt.Quote(file.FileName))
					}
				}
			} else if !fs.FileExists(fileName) {
				if opt.Dry {
					purgedFiles[fileName] = true
					log.Infof("purge: file %s would be removed", txt.Quote(file.FileName))
					continue
				}

				if err := file.Purge(); err != nil {
					log.Errorf("purge: %s", err)
				} else {
					purgedFiles[fileName] = true
					log.Infof("purge: removed file %s", txt.Quote(file.FileName))
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
				log.Infof("purge: photo %s would be removed", txt.Quote(photo.PhotoName))
				continue
			}

			if err := photo.Delete(opt.Hard); err != nil {
				log.Errorf("purge: %s", err)
			} else {
				purgedPhotos[photo.PhotoUID] = true

				if opt.Hard {
					log.Infof("purge: permanently deleted photo %s", txt.Quote(photo.PhotoName))
				} else {
					log.Infof("purge: removed photo %s", txt.Quote(photo.PhotoName))
				}
			}
		}

		if mutex.MainWorker.Canceled() {
			return purgedFiles, purgedPhotos, errors.New("purge canceled")
		}

		offset += limit

		time.Sleep(50 * time.Millisecond)
	}

	log.Info("purge: finding hidden photos")

	if err := query.ResetPhotoQuality(); err != nil {
		return purgedFiles, purgedPhotos, err
	}

	if err := entity.UpdatePhotoCounts(); err != nil {
		log.Errorf("purge: %s", err)
	}

	return purgedFiles, purgedPhotos, nil
}

// Cancel stops the current operation.
func (prg *Purge) Cancel() {
	mutex.MainWorker.Cancel()
}
