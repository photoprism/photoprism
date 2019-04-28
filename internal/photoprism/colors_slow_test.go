// +build slow

package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetColors_Slow(t *testing.T) {
	conf := test.NewConfig()

	conf.InitializeTestData(t)

	if mediaFile2, err := NewMediaFile(conf.ImportPath() + "/iphone/IMG_6788.JPG"); err == nil {
		colors, main, l, m, err := mediaFile2.Colors()

		t.Log(colors, main, l, m, err)

		assert.Nil(t, err)
		assert.False(t, m)
		assert.IsType(t, MaterialColors{}, colors)
		assert.Equal(t, "grey", main.Name())
		assert.Equal(t, MaterialColors{0x2, 0x1, 0x2, 0x1, 0x1, 0x1, 0x2, 0x1, 0x2}, colors)
	} else {
		t.Error(err)
	}

	if mediaFile3, err := NewMediaFile(conf.ImportPath() + "/raw/20140717_154212_1EC48F8489.jpg"); err == nil {
		colors, main, l, m, err := mediaFile3.Colors()

		t.Log(colors, main, l, m, err)

		assert.Nil(t, err)
		assert.False(t, m)
		assert.IsType(t, MaterialColors{}, colors)
		assert.Equal(t, "grey", main.Name())

		assert.Equal(t, MaterialColors{0x3, 0x2, 0x2, 0x1, 0x2, 0x2, 0x2, 0x2, 0x1}, colors)

	} else {
		t.Error(err)
	}
}
