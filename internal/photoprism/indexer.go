package photoprism

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/config"
)

// Indexer defines an indexer with originals path tensorflow and a db.
type Indexer struct {
	conf       *config.Config
	tensorFlow *TensorFlow
	db         *gorm.DB
}

// NewIndexer returns a new indexer.
// TODO: Is it really necessary to return a pointer?
func NewIndexer(conf *config.Config, tensorFlow *TensorFlow) *Indexer {
	instance := &Indexer{
		conf:       conf,
		tensorFlow: tensorFlow,
		db:         conf.Db(),
	}

	return instance
}

func (i *Indexer) originalsPath() string {
	return i.conf.OriginalsPath()
}

func (i *Indexer) thumbnailsPath() string {
	return i.conf.ThumbnailsPath()
}

// IndexRelated will index all mediafiles which has relate to a given mediafile.
func (i *Indexer) IndexRelated(mediaFile *MediaFile, o IndexerOptions) map[string]bool {
	indexed := make(map[string]bool)

	relatedFiles, mainFile, err := mediaFile.RelatedFiles()

	if err != nil {
		log.Warnf("could not index \"%s\": %s", mediaFile.RelativeFilename(i.originalsPath()), err.Error())

		return indexed
	}

	mainIndexResult := i.indexMediaFile(mainFile, o)
	indexed[mainFile.Filename()] = true

	log.Infof("index: %s main %s file \"%s\"", mainIndexResult, mainFile.Type(), mainFile.RelativeFilename(i.originalsPath()))

	for _, relatedMediaFile := range relatedFiles {
		if indexed[relatedMediaFile.Filename()] {
			continue
		}

		indexResult := i.indexMediaFile(relatedMediaFile, o)
		indexed[relatedMediaFile.Filename()] = true

		log.Infof("index: %s related %s file \"%s\"", indexResult, relatedMediaFile.Type(), relatedMediaFile.RelativeFilename(i.originalsPath()))
	}

	return indexed
}

// IndexOriginals will index mediafiles in the originals directory.
func (i *Indexer) IndexOriginals(o IndexerOptions) map[string]bool {
	indexed := make(map[string]bool)

	err := filepath.Walk(i.originalsPath(), func(filename string, fileInfo os.FileInfo, err error) error {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("index: panic %s", err)
			}
		}()
		if err != nil || indexed[filename] {
			return nil
		}

		if fileInfo.IsDir() || strings.HasPrefix(filepath.Base(filename), ".") {
			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !mediaFile.IsPhoto() {
			return nil
		}

		for relatedFilename := range i.IndexRelated(mediaFile, o) {
			indexed[relatedFilename] = true
		}

		return nil
	})

	if err != nil {
		log.Warn(err.Error())
	}

	return indexed
}
