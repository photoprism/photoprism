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

	indexer := NewIndexer(conf.GetOriginalsPath(), tensorFlow, conf.Db())

	converter := NewConverter(conf.GetDarktableCli())

	importer := NewImporter(conf.GetOriginalsPath(), indexer, converter)

	importer.ImportPhotosFromDirectory(conf.GetImportPath())
}
