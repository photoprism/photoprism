package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/nsfw"
)

func TestIndexer_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf)
	nsfwDetector := nsfw.NewDetector(conf.NSFWModelPath())

	indexer := NewIndexer(conf, tensorFlow, nsfwDetector)

	converter := NewConverter(conf)

	importer := NewImporter(conf, indexer, converter)

	importer.ImportPhotosFromDirectory(conf.ImportPath())

	options := IndexerOptionsAll()

	indexer.Start(options)
}
