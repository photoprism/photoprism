package photoprism

import (
	"testing"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestMediaFile_Thumbnail(t *testing.T) {
	conf := config.TestConfig()

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	conf.InitializeTestData(t)

	image, err := NewMediaFile(conf.ImportPath() + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)

	thumbnail, err := image.Thumbnail(conf.ThumbnailsPath(), "tile_500")

	assert.Empty(t, err)

	assert.FileExists(t, thumbnail)
}

func TestCreateThumbnailsFromOriginals(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf)

	indexer := NewIndexer(conf, tensorFlow)

	converter := NewConverter(conf)

	importer := NewImporter(conf, indexer, converter)

	importer.ImportPhotosFromDirectory(conf.ImportPath())

	err := CreateThumbnailsFromOriginals(conf.OriginalsPath(), conf.ThumbnailsPath(), true)

	if err != nil {
		t.Error(err)
	}
}

func TestResampleFile(t *testing.T) {
	conf := config.TestConfig()

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	conf.InitializeTestData(t)

	fileModel := &models.File{
		FileName: conf.ImportPath() + "/dog.jpg",
		FileHash: "123456789",
	}

	thumb, err := ThumbnailFromFile(fileModel.FileName, fileModel.FileHash, conf.ThumbnailsPath(), 224, 224)
	assert.Nil(t, err)

	assert.IsType(t, "", thumb)
}

func TestCreateThumbnail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	conf.InitializeTestData(t)

	expectedFilename, err := ThumbnailFilename("12345", conf.ThumbnailsPath(), 150, 150, ResampleFit, ResampleNearestNeighbor)

	if err != nil {
		t.Error(err)
	}

	img, err := imaging.Open(conf.ImportPath()+"/dog.jpg", imaging.AutoOrientation(true))

	if err != nil {
		t.Errorf("can't open original: %s", err)
	}

	thumb, err := CreateThumbnail(img, expectedFilename, 150, 150, ResampleFit, ResampleNearestNeighbor)

	assert.Empty(t, err)

	assert.NotNil(t, thumb)

	bounds := thumb.Bounds()

	assert.Equal(t, 150, bounds.Dx())
	assert.Equal(t, 106, bounds.Dy())

	assert.FileExists(t, expectedFilename)
}

func TestMediaFile_CreateDefaultThumbnails(t *testing.T) {
	conf := config.TestConfig()

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	conf.InitializeTestData(t)

	m, err := NewMediaFile(conf.ImportPath() + "/dog.jpg")
	assert.Nil(t, err)

	err = m.CreateDefaultThumbnails(conf.ThumbnailsPath(), true)

	assert.Empty(t, err)

	thumbFilename, err := ThumbnailFilename(m.Hash(), conf.ThumbnailsPath(), ThumbnailTypes["tile_50"].Width, ThumbnailTypes["tile_50"].Height, ThumbnailTypes["tile_50"].Options...)

	assert.Empty(t, err)

	assert.FileExists(t, thumbFilename)

	err = m.CreateDefaultThumbnails(conf.ThumbnailsPath(), false)

	assert.Empty(t, err)
}
