package photoprism

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMediaFile_GetColors(t *testing.T) {
	conf := NewTestConfig()

	conf.InitializeTestData(t)

	mediaFile1 := NewMediaFile(conf.ImportPath + "/dog.jpg")

	names, vibrantHex, mutedHex := mediaFile1.GetColors()

	t.Log(names, vibrantHex, mutedHex)

	assert.IsType(t, []string{}, names)
	assert.Equal(t, "#e0ed21", vibrantHex)
	assert.Equal(t, "#977d67", mutedHex)

	mediaFile2 := NewMediaFile(conf.ImportPath + "/iphone/IMG_6788.JPG")

	names, vibrantHex, mutedHex = mediaFile2.GetColors()

	t.Log(names, vibrantHex, mutedHex)

	assert.Equal(t, "#3d85c3", vibrantHex)
	assert.Equal(t, "#988570", mutedHex)

	mediaFile3 := NewMediaFile(conf.ImportPath + "/raw/20140717_154212_1EC48F8489.jpg")

	names, vibrantHex, mutedHex = mediaFile3.GetColors()

	t.Log(names, vibrantHex, mutedHex)

	assert.Equal(t, "#d5d437", vibrantHex)
	assert.Equal(t, "#a69f55", mutedHex)
}