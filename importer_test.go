package photoprism

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewImporter(t *testing.T) {
	conf := NewTestConfig()

	importer := NewImporter(conf.OriginalsPath)

	assert.IsType(t, &Importer{}, importer)
}

func TestImporter_ImportPhotosFromDirectory(t *testing.T) {
	conf := NewTestConfig()

	conf.InitializeTestData(t)

	importer := NewImporter(conf.OriginalsPath)

	importer.ImportPhotosFromDirectory(conf.ImportPath)
}

func TestImporter_GetDestinationFilename(t *testing.T) {
	conf := NewTestConfig()
	conf.InitializeTestData(t)
	importer := NewImporter(conf.OriginalsPath)

	rawFile := NewMediaFile(conf.ImportPath + "/raw/IMG_1435.cr2")

	filename, err := importer.GetDestinationFilename(rawFile, rawFile)

	assert.Empty(t, err)

	assert.Equal(t, conf.OriginalsPath + "/2018/02/20180204_170813_B0770443A5F7.cr2", filename)
}