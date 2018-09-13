package photoprism

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMediaFile_GetThumbnail(t *testing.T) {
	conf := NewTestConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	image1, err := NewMediaFile(conf.ImportPath + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)

	thumbnail1, err := image1.GetThumbnail(conf.ThumbnailsPath, 350)

	assert.Empty(t, err)

	assert.IsType(t, &MediaFile{}, thumbnail1)
}

func TestMediaFile_GetSquareThumbnail(t *testing.T) {
	conf := NewTestConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	image1, err := NewMediaFile(conf.ImportPath + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)

	thumbnail1, err := image1.GetSquareThumbnail(conf.ThumbnailsPath, 350)

	assert.Empty(t, err)

	assert.IsType(t, &MediaFile{}, thumbnail1)
}

func TestCreateThumbnailsFromOriginals(t *testing.T) {
	conf := NewTestConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	indexer := NewIndexer(conf.OriginalsPath, conf.GetDb())

	importer := NewImporter(conf.OriginalsPath, indexer)

	importer.ImportPhotosFromDirectory(conf.ImportPath)

	CreateThumbnailsFromOriginals(conf.OriginalsPath, conf.ThumbnailsPath, 600, false)

	CreateThumbnailsFromOriginals(conf.OriginalsPath, conf.ThumbnailsPath, 300, true)
}
