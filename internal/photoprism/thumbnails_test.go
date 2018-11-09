package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetThumbnail(t *testing.T) {
	conf := NewTestConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	image1, err := NewMediaFile(conf.GetImportPath() + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)

	thumbnail1, err := image1.GetThumbnail(conf.GetThumbnailsPath(), 350)

	assert.Empty(t, err)

	assert.IsType(t, &MediaFile{}, thumbnail1)
}

func TestMediaFile_GetSquareThumbnail(t *testing.T) {
	conf := NewTestConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	image1, err := NewMediaFile(conf.GetImportPath() + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)

	thumbnail1, err := image1.GetSquareThumbnail(conf.GetThumbnailsPath(), 350)

	assert.Empty(t, err)

	assert.IsType(t, &MediaFile{}, thumbnail1)
}

func TestCreateThumbnailsFromOriginals(t *testing.T) {
	conf := NewTestConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf.GetTensorFlowModelPath())

	indexer := NewIndexer(conf.GetOriginalsPath(), tensorFlow, conf.GetDb())

	converter := NewConverter(conf.GetDarktableCli())

	importer := NewImporter(conf.GetOriginalsPath(), indexer, converter)

	importer.ImportPhotosFromDirectory(conf.GetImportPath())

	CreateThumbnailsFromOriginals(conf.GetOriginalsPath(), conf.GetThumbnailsPath(), 600, false)

	CreateThumbnailsFromOriginals(conf.GetOriginalsPath(), conf.GetThumbnailsPath(), 300, true)
}
