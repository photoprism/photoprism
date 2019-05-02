package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetThumbnail(t *testing.T) {
	conf := test.NewConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	image1, err := NewMediaFile(conf.ImportPath() + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)

	thumbnail1, err := image1.Thumbnail(conf.ThumbnailsPath(), 350)

	assert.Empty(t, err)

	assert.IsType(t, &MediaFile{}, thumbnail1)
}

func TestMediaFile_GetSquareThumbnail(t *testing.T) {
	conf := test.NewConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	image1, err := NewMediaFile(conf.ImportPath() + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)

	thumbnail1, err := image1.SquareThumbnail(conf.ThumbnailsPath(), 350)

	assert.Empty(t, err)

	assert.IsType(t, &MediaFile{}, thumbnail1)
}

func TestCreateThumbnailsFromOriginals(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := test.NewConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf.TensorFlowModelPath())

	indexer := NewIndexer(conf.OriginalsPath(), tensorFlow, conf.Db())

	converter := NewConverter(conf.DarktableCli())

	importer := NewImporter(conf.OriginalsPath(), indexer, converter)

	importer.ImportPhotosFromDirectory(conf.ImportPath())

	CreateThumbnailsFromOriginals(conf.OriginalsPath(), conf.ThumbnailsPath(), 600, false)

	CreateThumbnailsFromOriginals(conf.OriginalsPath(), conf.ThumbnailsPath(), 300, true)
}
