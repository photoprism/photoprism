package photoprism

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"sync"

	"github.com/karrick/godirwalk"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Resample represents a thumbnail generator worker.
type Resample struct {
	conf *config.Config
}

// NewResample returns a new thumbnail generator and expects the config as argument.
func NewResample(conf *config.Config) *Resample {
	return &Resample{conf: conf}
}

// Start creates default thumbnails for all files in originalsPath.
func (w *Resample) Start(force bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("resample: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err := mutex.MainWorker.Start(); err != nil {
		return err
	}

	defer mutex.MainWorker.Stop()

	originalsPath := w.conf.OriginalsPath()
	thumbnailsPath := w.conf.ThumbPath()

	jobs := make(chan ResampleJob)

	// Start a fixed number of goroutines to read and digest files.
	var wg sync.WaitGroup
	var numWorkers = w.conf.Workers()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			ResampleWorker(jobs)
			wg.Done()
		}()
	}

	done := make(fs.Done)
	ignore := fs.NewIgnoreList(fs.IgnoreFile, true, false)

	if err := ignore.Dir(originalsPath); err != nil {
		log.Infof("resample: %s", err)
	}

	ignore.Log = func(fileName string) {
		log.Infof(`resample: ignored "%s"`, fs.RelName(fileName, originalsPath))
	}

	err = godirwalk.Walk(originalsPath, &godirwalk.Options{
		ErrorCallback: func(fileName string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
		Callback: func(fileName string, info *godirwalk.Dirent) error {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("resample: %s (panic)\nstack: %s", r, debug.Stack())
				}
			}()

			if mutex.MainWorker.Canceled() {
				return errors.New("canceled")
			}

			isDir := info.IsDir()
			isSymlink := info.IsSymlink()

			if skip, result := fs.SkipWalk(fileName, isDir, isSymlink, done, ignore); skip {
				return result
			}

			mf, err := NewMediaFile(fileName)

			if err != nil || !mf.IsJpeg() {
				return nil
			}

			done[fileName] = fs.Processed

			relativeName := mf.RelName(originalsPath)

			event.Publish("index.thumbnails", event.Data{
				"fileName": relativeName,
				"baseName": filepath.Base(relativeName),
				"force":    force,
			})

			jobs <- ResampleJob{
				mediaFile: mf,
				path:      thumbnailsPath,
				force:     force,
			}

			return nil
		},
		Unsorted:            true,
		FollowSymbolicLinks: true,
	})

	close(jobs)
	wg.Wait()

	return err
}
