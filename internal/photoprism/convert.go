package photoprism

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"sync"

	"github.com/karrick/godirwalk"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/list"
)

// Convert represents a file format conversion worker.
type Convert struct {
	conf               *config.Config
	cmdMutex           sync.Mutex
	sipsExclude        fs.ExtList
	darktableExclude   fs.ExtList
	rawTherapeeExclude fs.ExtList
	imageMagickExclude fs.ExtList
}

// NewConvert returns a new file format conversion worker.
func NewConvert(conf *config.Config) *Convert {
	c := &Convert{
		conf:               conf,
		sipsExclude:        fs.NewExtList(conf.SipsExclude()),
		darktableExclude:   fs.NewExtList(conf.DarktableExclude()),
		rawTherapeeExclude: fs.NewExtList(conf.RawTherapeeExclude()),
		imageMagickExclude: fs.NewExtList(conf.ImageMagickExclude()),
	}

	return c
}

// Start converts all files in the specified directory based on the current configuration.
func (w *Convert) Start(dir string, ext []string, force bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("convert: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err = mutex.IndexWorker.Start(); err != nil {
		return err
	}

	defer mutex.IndexWorker.Stop()

	jobs := make(chan ConvertJob)

	// Start a fixed number of goroutines to convert files.
	var wg sync.WaitGroup
	var numWorkers = w.conf.IndexWorkers()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			ConvertWorker(jobs)
			wg.Done()
		}()
	}

	done := make(fs.Done)
	ignore := fs.NewIgnoreList(fs.PPIgnoreFilename, true, false)

	if err = ignore.Path(dir); err != nil {
		log.Infof("convert: %s", err)
	}

	ignore.Log = func(fileName string) {
		log.Infof("convert: ignoring %s", clean.Log(filepath.Base(fileName)))
	}

	err = godirwalk.Walk(dir, &godirwalk.Options{
		ErrorCallback: func(fileName string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
		Callback: func(fileName string, info *godirwalk.Dirent) error {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("convert: %s (panic)\nstack: %s", r, debug.Stack())
				}
			}()

			if mutex.IndexWorker.Canceled() {
				return errors.New("canceled")
			}

			isDir, _ := info.IsDirOrSymlinkToDir()
			isSymlink := info.IsSymlink()

			// Skip file?
			if skip, result := fs.SkipWalk(fileName, isDir, isSymlink, done, ignore); skip {
				return result
			}

			// Process only files with specified extensions?
			if list.Excludes(ext, fs.NormalizedExt(fileName)) {
				return nil
			}

			f, err := NewMediaFile(fileName)

			if err != nil || f.Empty() || f.IsPreviewImage() || !f.IsMedia() {
				return nil
			}

			done[fileName] = fs.Processed

			jobs <- ConvertJob{
				force:   force,
				file:    f,
				convert: w,
			}

			return nil
		},
		Unsorted:            false,
		FollowSymbolicLinks: true,
	})

	close(jobs)
	wg.Wait()

	return err
}
