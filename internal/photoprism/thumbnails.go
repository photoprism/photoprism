package photoprism

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
)

// Thumbnails represents a thumbnail generator.
type Thumbnails struct {
	conf *config.Config
}

// NewThumbnails returns a new thumbnail generator and expects the config as argument.
func NewThumbnails(conf *config.Config) *Thumbnails {
	return &Thumbnails{conf: conf}
}

// Start creates default thumbnails for all files in originalsPath.
func (t *Thumbnails) Start(force bool) error {
	if err := mutex.Worker.Start(); err != nil {
		return err
	}

	defer mutex.Worker.Stop()

	originalsPath := t.conf.OriginalsPath()
	thumbnailsPath := t.conf.ThumbnailsPath()

	jobs := make(chan ThumbnailsJob)

	// Start a fixed number of goroutines to read and digest files.
	var wg sync.WaitGroup
	var numWorkers = t.conf.Workers()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			thumbnailsWorker(jobs) // HLc
			wg.Done()
		}()
	}

	err := filepath.Walk(originalsPath, func(filename string, fileInfo os.FileInfo, err error) error {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("thumbs: %s [panic]", err)
			}
		}()

		if mutex.Worker.Canceled() {
			return errors.New("thumbs: canceled")
		}

		if err != nil || fileInfo.IsDir() || strings.HasPrefix(filepath.Base(filename), ".") {
			return nil
		}

		mf, err := NewMediaFile(filename)

		if err != nil || !mf.IsJpeg() {
			return nil
		}

		fileName := mf.RelativeFilename(originalsPath)

		event.Publish("index.thumbnails", event.Data{
			"fileName": fileName,
			"baseName": filepath.Base(fileName),
			"force":    force,
		})

		jobs <- ThumbnailsJob{
			mediaFile: mf,
			path:      thumbnailsPath,
			force:     force,
		}

		return nil
	})

	close(jobs)
	wg.Wait()

	return err
}
