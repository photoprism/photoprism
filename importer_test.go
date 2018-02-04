package photoprism

import (
	"testing"
)

func TestImporter_ImportFromDirectory(t *testing.T) {
	config := NewTestConfig()

	converter := NewConverter(config.DarktableCli)
	importer := NewImporter(config.OriginalsPath, converter)

	importer.CreateJpegFromRaw(config.ImportPath)
	importer.ImportJpegFromDirectory(config.ImportPath)
}