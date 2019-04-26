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
		colors, main, err := mediaFile1.Colors()

		t.Log(colors, main, err)

		assert.Nil(t, err)
		assert.IsType(t, MaterialColors{}, colors)
		assert.Equal(t, "grey", main.Name())
		assert.Equal(t, MaterialColors{0x1, 0x2, 0x1, 0x2, 0x2, 0x1, 0x1, 0x1, 0x0}, colors)
	} else {
		t.Error(err)
	}

	if mediaFile2, err := NewMediaFile(conf.ImportPath() + "/ape.jpeg"); err == nil {
		colors, main, err := mediaFile2.Colors()

		t.Log(colors, main, err)

		assert.Nil(t, err)
		assert.IsType(t, MaterialColors{}, colors)
		assert.Equal(t, "teal", main.Name())
		assert.Equal(t, MaterialColors{0x8, 0x8, 0x2, 0x8, 0x2, 0x1, 0x8, 0x1, 0x2}, colors)
	} else {
		t.Error(err)
	}
}
