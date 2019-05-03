package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/context"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetThumbnail(t *testing.T) {
	ctx := context.TestContext()

	ctx.CreateDirectories()

	ctx.InitializeTestData(t)

	image1, err := NewMediaFile(ctx.ImportPath() + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)

	thumbnail1, err := image1.Thumbnail(ctx.ThumbnailsPath(), 350)

	assert.Empty(t, err)

	assert.IsType(t, &MediaFile{}, thumbnail1)
}

func TestMediaFile_GetSquareThumbnail(t *testing.T) {
	ctx := context.TestContext()

	ctx.CreateDirectories()

	ctx.InitializeTestData(t)

	image1, err := NewMediaFile(ctx.ImportPath() + "/iphone/IMG_6788.JPG")
	assert.Nil(t, err)

	thumbnail1, err := image1.SquareThumbnail(ctx.ThumbnailsPath(), 350)

	assert.Empty(t, err)

	assert.IsType(t, &MediaFile{}, thumbnail1)
}

func TestCreateThumbnailsFromOriginals(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.TestContext()

	ctx.CreateDirectories()

	ctx.InitializeTestData(t)

	tensorFlow := NewTensorFlow(ctx.TensorFlowModelPath())

	indexer := NewIndexer(ctx.OriginalsPath(), tensorFlow, ctx.Db())

	converter := NewConverter(ctx.DarktableCli())

	importer := NewImporter(ctx.OriginalsPath(), indexer, converter)

	importer.ImportPhotosFromDirectory(ctx.ImportPath())

	CreateThumbnailsFromOriginals(ctx.OriginalsPath(), ctx.ThumbnailsPath(), 600, false)

	CreateThumbnailsFromOriginals(ctx.OriginalsPath(), ctx.ThumbnailsPath(), 300, true)
}
