package photoprism

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/nsfw"
)

// Indexer defines an indexer with originals path tensorflow and a db.
type Indexer struct {
	conf         *config.Config
	tensorFlow   *TensorFlow
	nsfwDetector *nsfw.Detector
	db           *gorm.DB
	running      bool
	canceled     bool
}

// NewIndexer returns a new indexer.
// TODO: Is it really necessary to return a pointer?
func NewIndexer(conf *config.Config, tensorFlow *TensorFlow, nsfwDetector *nsfw.Detector) *Indexer {
	i := &Indexer{
		conf:         conf,
		tensorFlow:   tensorFlow,
		nsfwDetector: nsfwDetector,
		db:           conf.Db(),
	}

	return i
}

func (ind *Indexer) originalsPath() string {
	return ind.conf.OriginalsPath()
}

func (ind *Indexer) thumbnailsPath() string {
	return ind.conf.ThumbnailsPath()
}

// Cancel stops the current indexing operation.
func (ind *Indexer) Cancel() {
	ind.canceled = true
}

// Start will index mediafiles in the originals directory.
func (ind *Indexer) Start(options IndexerOptions) map[string]bool {
	done := make(map[string]bool)

	if ind.running {
		event.Error("indexer already running")
		return done
	}

	ind.running = true
	ind.canceled = false

	defer func() {
		ind.running = false
		ind.canceled = false
	}()

	if err := ind.tensorFlow.Init(); err != nil {
		log.Errorf("index: %s", err.Error())

		return done
	}

	jobs := make(chan IndexJob)

	// Start a fixed number of goroutines to read and digest files.
	var wg sync.WaitGroup
	var numWorkers = ind.conf.Workers()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			indexerWorker(jobs) // HLc
			wg.Done()
		}()
	}

	err := filepath.Walk(ind.originalsPath(), func(filename string, fileInfo os.FileInfo, err error) error {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("index: panic %s", err)
			}
		}()

		if ind.canceled {
			return errors.New("indexing canceled")
		}

		if err != nil || done[filename] {
			return nil
		}

		if fileInfo.IsDir() || strings.HasPrefix(filepath.Base(filename), ".") {
			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !mediaFile.IsPhoto() {
			return nil
		}

		related, err := mediaFile.RelatedFiles()

		if err != nil {
			log.Warnf("could not index \"%s\": %s", mediaFile.RelativeFilename(ind.originalsPath()), err.Error())

			return nil
		}

		for _, f := range related.files {
			done[f.Filename()] = true
		}

		jobs <- IndexJob{
			related: related,
			options: options,
			ind:     ind,
		}

		return nil
	})

	close(jobs)
	wg.Wait()

	if err != nil {
		log.Error(err.Error())
	}

	return done
}
