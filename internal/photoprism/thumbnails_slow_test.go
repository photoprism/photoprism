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

	indexer := NewIndexer(conf.OriginalsPath(), tensorFlow, conf.Db())

	converter := NewConverter(conf.GetDarktableCli())

	importer := NewImporter(conf.OriginalsPath(), indexer, converter)

	importer.ImportPhotosFromDirectory(conf.GetImportPath())

	CreateThumbnailsFromOriginals(conf.OriginalsPath(), conf.GetThumbnailsPath(), 600, false)

	CreateThumbnailsFromOriginals(conf.OriginalsPath(), conf.GetThumbnailsPath(), 300, true)
}
