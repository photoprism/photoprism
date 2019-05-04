package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/context"
	"github.com/stretchr/testify/assert"
)

func TestNewImporter(t *testing.T) {
	ctx := context.TestContext()

	tensorFlow := NewTensorFlow(ctx.TensorFlowModelPath())

	indexer := NewIndexer(ctx.OriginalsPath(), tensorFlow, ctx.Db())

	converter := NewConverter(ctx.DarktableCli())

	importer := NewImporter(ctx.OriginalsPath(), indexer, converter)

	assert.IsType(t, &Importer{}, importer)
}

func TestImporter_DestinationFilename(t *testing.T) {
	ctx := context.TestContext()

	ctx.InitializeTestData(t)

	tensorFlow := NewTensorFlow(ctx.TensorFlowModelPath())

	indexer := NewIndexer(ctx.OriginalsPath(), tensorFlow, ctx.Db())

	converter := NewConverter(ctx.DarktableCli())

	importer := NewImporter(ctx.OriginalsPath(), indexer, converter)

	rawFile, err := NewMediaFile(ctx.ImportPath() + "/raw/IMG_1435.CR2")

	assert.Nil(t, err)

	filename, err := importer.DestinationFilename(rawFile, rawFile)

	assert.Nil(t, err)

	assert.Equal(t, ctx.OriginalsPath()+"/2018/02/20180204_170813_863A6248DCCA.cr2", filename)
}

func TestImporter_ImportPhotosFromDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.TestContext()

	ctx.InitializeTestData(t)

	tensorFlow := NewTensorFlow(ctx.TensorFlowModelPath())

	indexer := NewIndexer(ctx.OriginalsPath(), tensorFlow, ctx.Db())

	converter := NewConverter(ctx.DarktableCli())

	importer := NewImporter(ctx.OriginalsPath(), indexer, converter)

	importer.ImportPhotosFromDirectory(ctx.ImportPath())
}
