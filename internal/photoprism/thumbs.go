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
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Thumbs represents a thumbnail image generator.
type Thumbs struct {
	conf *config.Config
}

// NewThumbs returns a new thumbnails generator and expects the config as argument.
func NewThumbs(conf *config.Config) *Thumbs {
	return &Thumbs{conf: conf}
}

// Start creates thumbnails for files in the originals and sidecar folders.
func (w *Thumbs) Start(dir string, force, originalsOnly bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("thumbs: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	originalsPath := w.conf.OriginalsPath()
	originalsDir := filepath.Join(originalsPath, dir)
	sidecarPath := w.conf.SidecarPath()
	sidecarDir := filepath.Join(sidecarPath, dir)

	// Valid path provided?
	if !fs.PathExists(originalsDir) {
		return fmt.Errorf("thumbs: directory %s not found", clean.Log(originalsDir))
	}

	// Scan sidecar folder?
	originalsOnly = originalsOnly || sidecarPath == "" || sidecarPath == originalsPath || !fs.PathExists(sidecarDir)

	// Start creating thumbnails.
	if _, err = w.Dir(originalsDir, force); err != nil || originalsOnly {
		return err
	} else if _, err = w.Dir(sidecarDir, force); err != nil {
		return err
	}

	return nil
}

// Dir creates thumbnail images for files found in a given path.
func (w *Thumbs) Dir(dir string, force bool) (fs.Done, error) {
	done := make(fs.Done)

	if err := mutex.MainWorker.Start(); err != nil {
		return done, err
	}

	defer mutex.MainWorker.Stop()

	jobs := make(chan ThumbsJob)
	thumbnailsPath := w.conf.ThumbCachePath()

	// Start a fixed number of goroutines to read and digest files.
	var wg sync.WaitGroup
	var numWorkers = w.conf.Workers()

	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			ThumbsWorker(jobs)
			wg.Done()
		}()
	}

	ignore := fs.NewIgnoreList(fs.PPIgnoreFilename, true, false)
	ignore.Log = func(fileName string) {
		log.Infof(`thumbs: ignored "%s"`, fs.RelName(fileName, dir))
	}

	handler := func(fileName string, info *godirwalk.Dirent) error {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("thumbs: %s (panic)\nstack: %s", r, debug.Stack())
			}
		}()

		if mutex.MainWorker.Canceled() {
			return errors.New("canceled")
		}

		isDir, _ := info.IsDirOrSymlinkToDir()
		isSymlink := info.IsSymlink()

		if skip, result := fs.SkipWalk(fileName, isDir, isSymlink, done, ignore); skip {
			return result
		}

		mf, err := NewMediaFile(fileName)

		if err != nil || mf.Empty() || !mf.IsPreviewImage() {
			return nil
		}

		done[fileName] = fs.Processed

		relativeName := mf.RelName(dir)

		event.Publish("index.thumbnails", event.Data{
			"fileName": relativeName,
			"baseName": filepath.Base(relativeName),
			"force":    force,
		})

		jobs <- ThumbsJob{
			mediaFile: mf,
			path:      thumbnailsPath,
			force:     force,
		}

		return nil
	}

	log.Infof("thumbs: processing %s", clean.Log(dir))

	if err := ignore.Dir(dir); err != nil {
		log.Infof("thumbs: %s", err)
	}

	err := godirwalk.Walk(dir, &godirwalk.Options{
		ErrorCallback: func(fileName string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
		Callback:            handler,
		Unsorted:            false,
		FollowSymbolicLinks: true,
	})

	close(jobs)
	wg.Wait()

	return done, err
}
