// +build slow

package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/test"
)

func TestCreateThumbnailsFromOriginals(t *testing.T) {
	conf := test.NewConfig()

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
