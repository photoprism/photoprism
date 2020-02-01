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

// Resample represents a thumbnail generator.
type Resample struct {
	conf *config.Config
}

// NewResample returns a new thumbnail generator and expects the config as argument.
func NewResample(conf *config.Config) *Resample {
	return &Resample{conf: conf}
}

// Start creates default thumbnails for all files in originalsPath.
func (rs *Resample) Start(force bool) error {
	if err := mutex.Worker.Start(); err != nil {
		return err
	}

	defer mutex.Worker.Stop()

	originalsPath := rs.conf.OriginalsPath()
	thumbnailsPath := rs.conf.ThumbnailsPath()

	jobs := make(chan ResampleJob)

	// Start a fixed number of goroutines to read and digest files.
	var wg sync.WaitGroup
	var numWorkers = rs.conf.Workers()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			resampleWorker(jobs)
			wg.Done()
		}()
	}

	err := filepath.Walk(originalsPath, func(filename string, fileInfo os.FileInfo, err error) error {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("resample: %s [panic]", err)
			}
		}()

		if mutex.Worker.Canceled() {
			return errors.New("resample: canceled")
		}

		if err != nil || fileInfo.IsDir() || strings.HasPrefix(filepath.Base(filename), ".") {
			return nil
		}

		mf, err := NewMediaFile(filename)

		if err != nil || !mf.IsJpeg() {
			return nil
		}

		fileName := mf.RelativeName(originalsPath)

		event.Publish("index.thumbnails", event.Data{
			"fileName": fileName,
			"baseName": filepath.Base(fileName),
			"force":    force,
		})

		jobs <- ResampleJob{
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
