package photoprism

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
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

	originalsPath := w.conf.OriginalsPath()

	// Check if originals folder is empty.
	if fs.DirIsEmpty(originalsPath) {
		return purgedFiles, purgedPhotos, err
	}

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
						log.Infof("purge: found %s", clean.Log(file.FileName))
						continue
					}

					if err := file.Found(); err != nil {
						log.Errorf("purge: %s", err)
					} else {
						log.Infof("purge: found %s", clean.Log(file.FileName))
					}
				}
			} else if !fs.FileExists(fileName) {
				if opt.Dry {
					purgedFiles[fileName] = true
					log.Infof("purge: file %s would be flagged as missing", clean.Log(file.FileName))
					continue
				}

				wasPrimary := file.FilePrimary

				if err := file.Purge(); err != nil {
					log.Errorf("purge: %s", err)
					continue
				}

				w.files.Remove(file.FileName, file.FileRoot)
				purgedFiles[fileName] = true
				log.Infof("purge: flagged file %s as missing", clean.Log(file.FileName))

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

		time.Sleep(10 * time.Millisecond)
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
					log.Infof("purge: duplicate %s would be removed from index", clean.Log(file.FileName))
					continue
				}

				if err := file.Purge(); err != nil {
					log.Errorf("purge: %s", err)
				} else {
					w.files.Remove(file.FileName, file.FileRoot)
					purgedFiles[fileName] = true
					log.Infof("purge: removed duplicate %s from index", clean.Log(file.FileName))
				}
			}
		}

		if mutex.MainWorker.Canceled() {
			return purgedFiles, purgedPhotos, errors.New("purge canceled")
		}

		offset += limit

		time.Sleep(10 * time.Millisecond)
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
				log.Infof("purge: %s would be removed", photo.String())
				continue
			}

			if files, err := photo.Delete(opt.Hard); err != nil {
				log.Errorf("purge: %s (delete photo)", err)
			} else {
				purgedPhotos[photo.PhotoUID] = true

				if opt.Hard {
					log.Infof("purge: permanently removed %s", photo.String())
				} else {
					log.Infof("purge: flagged photo %s as deleted", photo.String())
				}

				// Remove files from lookup table.
				for _, file := range files {
					w.files.Remove(file.FileName, file.FileRoot)
				}
			}
		}

		if mutex.MainWorker.Canceled() {
			return purgedFiles, purgedPhotos, errors.New("purge canceled")
		}

		offset += limit

		time.Sleep(10 * time.Millisecond)
	}

	// Skip the index update if there are no changes.
	if len(purgedFiles) == 0 && len(purgedPhotos) == 0 && !opt.Force {
		return purgedFiles, purgedPhotos, nil
	}

	if err = query.FixPrimaries(); err != nil {
		log.Errorf("index: %s (update primary files)", err)
	}

	// Set photo quality scores to -1 if files are missing.
	if err = query.FlagHiddenPhotos(); err != nil {
		return purgedFiles, purgedPhotos, err
	}

	// Remove orphan index entries.
	if opt.Dry {
		if files, err := query.OrphanFiles(); err != nil {
			log.Errorf("index: %s (find orphan files)", err)
		} else if l := len(files); l > 0 {
			log.Infof("index: found %s", english.Plural(l, "orphan file", "orphan files"))
		} else {
			log.Infof("index: found no orphan files")
		}
	} else {
		if err = query.PurgeOrphans(); err != nil {
			log.Errorf("index: %s (purge orphans)", err)
		}

		// Regenerate search index columns.
		entity.File{}.RegenerateIndex()
	}

	// Hide missing album contents.
	if err = query.UpdateMissingAlbumEntries(); err != nil {
		log.Errorf("index: %s (update album entries)", err)
	}

	// Remove unused entries from the places table.
	if err = query.PurgePlaces(); err != nil {
		log.Errorf("index: %s (purge places)", err)
	}

	// Update precalculated photo and file counts.
	if err = entity.UpdateCounts(); err != nil {
		log.Warnf("index: %s (update counts)", err)
	}

	// Update album, subject, and label cover thumbs.
	if err = query.UpdateCovers(); err != nil {
		log.Warnf("index: %s (update covers)", err)
	}

	return purgedFiles, purgedPhotos, nil
}

// Cancel stops the current operation.
func (w *Purge) Cancel() {
	mutex.MainWorker.Cancel()
}
