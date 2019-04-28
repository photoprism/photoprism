package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetColors(t *testing.T) {
	conf := test.NewConfig()

	conf.InitializeTestData(t)

	if mediaFile1, err := NewMediaFile(conf.ImportPath() + "/dog.jpg"); err == nil {
		colors, main, l, m, err := mediaFile1.Colors()

		t.Log(colors, main, l, m, err)

		assert.Nil(t, err)
		assert.False(t, m)
		assert.IsType(t, MaterialColors{}, colors)
		assert.Equal(t, "grey", main.Name())
		assert.Equal(t, MaterialColors{0x1, 0x2, 0x1, 0x2, 0x2, 0x1, 0x1, 0x1, 0x0}, colors)
		assert.Equal(t, LightMap{5, 9, 7, 10, 9, 5, 5, 6, 2}, l)
	} else {
		t.Error(err)
	}

	if mediaFile2, err := NewMediaFile(conf.ImportPath() + "/ape.jpeg"); err == nil {
		colors, main, l, m, err := mediaFile2.Colors()

		t.Log(colors, main, l, m, err)

		assert.Nil(t, err)
		assert.False(t, m)
		assert.IsType(t, MaterialColors{}, colors)
		assert.Equal(t, "teal", main.Name())
		assert.Equal(t, MaterialColors{0x8, 0x8, 0x2, 0x8, 0x2, 0x1, 0x8, 0x1, 0x2}, colors)
		assert.Equal(t, LightMap{8, 8, 7, 7, 7, 5, 8, 6, 8}, l)
	} else {
		t.Error(err)
	}
}
