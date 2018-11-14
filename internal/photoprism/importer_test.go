package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewImporter(t *testing.T) {
	conf := NewTestConfig()

	tensorFlow := NewTensorFlow(conf.GetTensorFlowModelPath())

	indexer := NewIndexer(conf.GetOriginalsPath(), tensorFlow, conf.GetDb())

	converter := NewConverter(conf.GetDarktableCli())

	importer := NewImporter(conf.GetOriginalsPath(), indexer, converter)

	assert.IsType(t, &Importer{}, importer)
}

func TestImporter_GetDestinationFilename(t *testing.T) {
	conf := NewTestConfig()
	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf.GetTensorFlowModelPath())

	indexer := NewIndexer(conf.GetOriginalsPath(), tensorFlow, conf.GetDb())

	converter := NewConverter(conf.GetDarktableCli())

	importer := NewImporter(conf.GetOriginalsPath(), indexer, converter)

	rawFile, err := NewMediaFile(conf.GetImportPath() + "/raw/IMG_1435.CR2")

	assert.Nil(t, err)

	filename, err := importer.GetDestinationFilename(rawFile, rawFile)

	assert.Nil(t, err)

	assert.Equal(t, conf.GetOriginalsPath()+"/2018/02/20180204_170813_863A6248DCCA.cr2", filename)
}
