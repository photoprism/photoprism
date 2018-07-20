package photoprism

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewImporter(t *testing.T) {
	conf := NewTestConfig()

	indexer := NewIndexer(conf.OriginalsPath, conf.GetDb())

	importer := NewImporter(conf.OriginalsPath, indexer)

	assert.IsType(t, &Importer{}, importer)
}

func TestImporter_ImportPhotosFromDirectory(t *testing.T) {
	conf := NewTestConfig()

	conf.InitializeTestData(t)

	indexer := NewIndexer(conf.OriginalsPath, conf.GetDb())

	importer := NewImporter(conf.OriginalsPath, indexer)

	importer.ImportPhotosFromDirectory(conf.ImportPath)
}

func TestImporter_GetDestinationFilename(t *testing.T) {
	conf := NewTestConfig()
	conf.InitializeTestData(t)

	indexer := NewIndexer(conf.OriginalsPath, conf.GetDb())

	importer := NewImporter(conf.OriginalsPath, indexer)

	rawFile := NewMediaFile(conf.ImportPath + "/raw/IMG_1435.CR2")

	filename, err := importer.GetDestinationFilename(rawFile, rawFile)

	assert.Empty(t, err)

	assert.Equal(t, conf.OriginalsPath + "/2018/02/20180204_170813_B0770443A5F7.cr2", filename)
}
