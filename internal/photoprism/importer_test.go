package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/nsfw"
	"github.com/stretchr/testify/assert"
)

func TestNewImporter(t *testing.T) {
	conf := config.TestConfig()

	tensorFlow := NewTensorFlow(conf)
	nsfwDetector := nsfw.NewDetector(conf.NSFWModelPath())

	indexer := NewIndexer(conf, tensorFlow, nsfwDetector)

	converter := NewConverter(conf)

	importer := NewImporter(conf, indexer, converter)

	assert.IsType(t, &Importer{}, importer)
}

func TestImporter_DestinationFilename(t *testing.T) {
	conf := config.TestConfig()

	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf)
	nsfwDetector := nsfw.NewDetector(conf.NSFWModelPath())

	indexer := NewIndexer(conf, tensorFlow, nsfwDetector)

	converter := NewConverter(conf)

	importer := NewImporter(conf, indexer, converter)

	rawFile, err := NewMediaFile(conf.ImportPath() + "/raw/IMG_2567.CR2")

	assert.Nil(t, err)

	filename, _ := importer.DestinationFilename(rawFile, rawFile)

	// TODO: Check for errors!

	assert.Equal(t, conf.OriginalsPath()+"/2019/07/20190705_153230_6E16EB388AD2.cr2", filename)
}

func TestImporter_Start(t *testing.T) {
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

	importer.Start(conf.ImportPath())
}
