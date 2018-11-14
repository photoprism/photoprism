// +build slow

package photoprism

import (
	"testing"
)

func TestImporter_ImportPhotosFromDirectory(t *testing.T) {
	conf := NewTestConfig()

	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf.GetTensorFlowModelPath())

	indexer := NewIndexer(conf.GetOriginalsPath(), tensorFlow, conf.GetDb())

	converter := NewConverter(conf.GetDarktableCli())

	importer := NewImporter(conf.GetOriginalsPath(), indexer, converter)

	importer.ImportPhotosFromDirectory(conf.GetImportPath())
}