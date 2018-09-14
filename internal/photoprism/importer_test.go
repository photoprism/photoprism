package photoprism

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewImporter(t *testing.T) {
	conf := NewTestConfig()

	tensorFlow := NewTensorFlow(conf.GetTensorFlowModelPath())

	indexer := NewIndexer(conf.OriginalsPath, tensorFlow, conf.GetDb())

	importer := NewImporter(conf.OriginalsPath, indexer)

	assert.IsType(t, &Importer{}, importer)
}

func TestImporter_ImportPhotosFromDirectory(t *testing.T) {
	conf := NewTestConfig()

	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf.GetTensorFlowModelPath())

	indexer := NewIndexer(conf.OriginalsPath, tensorFlow, conf.GetDb())

	importer := NewImporter(conf.OriginalsPath, indexer)

	importer.ImportPhotosFromDirectory(conf.ImportPath)
}

func TestImporter_GetDestinationFilename(t *testing.T) {
	conf := NewTestConfig()
	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf.GetTensorFlowModelPath())

	indexer := NewIndexer(conf.OriginalsPath, tensorFlow, conf.GetDb())

	importer := NewImporter(conf.OriginalsPath, indexer)

	rawFile, err := NewMediaFile(conf.ImportPath + "/raw/IMG_1435.CR2")

	assert.Nil(t, err)

	filename, err := importer.GetDestinationFilename(rawFile, rawFile)

	assert.Nil(t, err)

	assert.Equal(t, conf.OriginalsPath+"/2018/02/20180204_170813_B0770443A5F7.cr2", filename)
}
