package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewImporter(t *testing.T) {
	conf := config.TestConfig()

	tensorFlow := NewTensorFlow(conf)

	indexer := NewIndexer(conf, tensorFlow)

	converter := NewConverter(conf)

	importer := NewImporter(conf, indexer, converter)

	assert.IsType(t, &Importer{}, importer)
}

func TestImporter_DestinationFilename(t *testing.T) {
	conf := config.TestConfig()

	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf)

	indexer := NewIndexer(conf, tensorFlow)

	converter := NewConverter(conf)

	importer := NewImporter(conf, indexer, converter)

	rawFile, err := NewMediaFile(conf.ImportPath() + "/raw/IMG_1435.CR2")

	assert.Nil(t, err)

	filename, err := importer.DestinationFilename(rawFile, rawFile)

	assert.Nil(t, err)

	assert.Equal(t, conf.OriginalsPath()+"/2018/02/20180204_170813_863A6248DCCA.cr2", filename)
}

func TestImporter_ImportPhotosFromDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf)

	indexer := NewIndexer(conf, tensorFlow)

	converter := NewConverter(conf)

	importer := NewImporter(conf, indexer, converter)

	importer.ImportPhotosFromDirectory(conf.ImportPath())
}
