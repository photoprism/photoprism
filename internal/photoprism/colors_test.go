package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetColors(t *testing.T) {
	conf := NewTestConfig()

	conf.InitializeTestData(t)

	if mediaFile1, err := NewMediaFile(conf.ImportPath + "/dog.jpg"); err == nil {
		names, vibrantHex, mutedHex := mediaFile1.GetColors()

		t.Log(names, vibrantHex, mutedHex)

		assert.IsType(t, []string{}, names)
		assert.Equal(t, "#e0ed21", vibrantHex)
		assert.Equal(t, "#977d67", mutedHex)
	} else {
		t.Error(err)
	}

	if mediaFile2, err := NewMediaFile(conf.ImportPath + "/iphone/IMG_6788.JPG"); err == nil {

		names, vibrantHex, mutedHex := mediaFile2.GetColors()

		t.Log(names, vibrantHex, mutedHex)

		assert.Equal(t, "#3d85c3", vibrantHex)
		assert.Equal(t, "#988570", mutedHex)
	} else {
		t.Error(err)
	}

	if mediaFile3, err := NewMediaFile(conf.ImportPath + "/raw/20140717_154212_1EC48F8489.jpg"); err == nil {

		names, vibrantHex, mutedHex := mediaFile3.GetColors()

		t.Log(names, vibrantHex, mutedHex)

		assert.Equal(t, "#d5d437", vibrantHex)
		assert.Equal(t, "#a69f55", mutedHex)
	} else {
		t.Error(err)
	}
}
