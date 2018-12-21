// +build slow

package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/test"
)

func TestImporter_ImportPhotosFromDirectory(t *testing.T) {
	conf := test.NewConfig()

	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf.GetTensorFlowModelPath())

	indexer := NewIndexer(conf.OriginalsPath(), tensorFlow, conf.Db())

	converter := NewConverter(conf.DarktableCli())

	importer := NewImporter(conf.OriginalsPath(), indexer, converter)

	importer.ImportPhotosFromDirectory(conf.ImportPath())
}
