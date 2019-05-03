package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/context"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetColors(t *testing.T) {
	ctx := context.TestContext()

	ctx.InitializeTestData(t)

	t.Run("dog.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(ctx.ImportPath() + "/dog.jpg"); err == nil {
			colors, main, l, s, err := mediaFile.Colors()

			t.Log(colors, main, l, s, err)

			assert.Nil(t, err)
			assert.Equal(t, 3, s.Int())
			assert.IsType(t, MaterialColors{}, colors)
			assert.Equal(t, "grey", main.Name())
			assert.Equal(t, MaterialColors{0x1, 0x2, 0x1, 0x2, 0x2, 0x1, 0x1, 0x1, 0x0}, colors)
			assert.Equal(t, LightMap{5, 9, 7, 10, 9, 5, 5, 6, 2}, l)
		} else {
			t.Error(err)
		}
	})

	t.Run("ape.jpeg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(ctx.ImportPath() + "/ape.jpeg"); err == nil {
			colors, main, l, s, err := mediaFile.Colors()

			t.Log(colors, main, l, s, err)

			assert.Nil(t, err)
			assert.Equal(t, 3, s.Int())
			assert.IsType(t, MaterialColors{}, colors)
			assert.Equal(t, "teal", main.Name())
			assert.Equal(t, MaterialColors{0x8, 0x8, 0x2, 0x8, 0x2, 0x1, 0x8, 0x1, 0x2}, colors)
			assert.Equal(t, LightMap{8, 8, 7, 7, 7, 5, 8, 6, 8}, l)
		} else {
			t.Error(err)
		}
	})

	t.Run("iphone/IMG_6788.JPG", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(ctx.ImportPath() + "/iphone/IMG_6788.JPG"); err == nil {
			colors, main, l, s, err := mediaFile.Colors()

			t.Log(colors, main, l, s, err)

			assert.Nil(t, err)
			assert.Equal(t, 3, s.Int())
			assert.IsType(t, MaterialColors{}, colors)
			assert.Equal(t, "grey", main.Name())
			assert.Equal(t, MaterialColors{0x2, 0x1, 0x2, 0x1, 0x1, 0x1, 0x2, 0x1, 0x2}, colors)
		} else {
			t.Error(err)
		}
	})

	t.Run("raw/20140717_154212_1EC48F8489.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(ctx.ImportPath() + "/raw/20140717_154212_1EC48F8489.jpg"); err == nil {
			colors, main, l, s, err := mediaFile.Colors()

			t.Log(colors, main, l, s, err)

			assert.Nil(t, err)
			assert.Equal(t, 2, s.Int())
			assert.IsType(t, MaterialColors{}, colors)
			assert.Equal(t, "grey", main.Name())

			assert.Equal(t, MaterialColors{0x3, 0x2, 0x2, 0x1, 0x2, 0x2, 0x2, 0x2, 0x1}, colors)
		} else {
			t.Error(err)
		}
	})
}
