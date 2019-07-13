package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_Colors_Testdata(t *testing.T) {
	conf := config.TestConfig()

	thumbsPath := conf.CachePath() + "/_tmp"

	defer os.RemoveAll(thumbsPath)

	/*
		TODO: Add and compare other images in "testdata/"
	*/
	expected := map[string]ColorPerception{
		"elephant_mono.jpg": {
			Colors:    IndexedColors{0x2, 0x2, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0},
			MainColor: 0,
			Luminance: LightMap{0xa, 0x9, 0x0, 0x0, 0x6, 0x0, 0x0, 0x0, 0x0},
			Chroma:    0,
		},
		"sharks_blue.jpg": {
			Colors:    IndexedColors{0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x4, 0x4, 0x6},
			MainColor: 6,
			Luminance: LightMap{0x9, 0x7, 0x5, 0x4, 0x3, 0x4, 0x3, 0x3, 0x3},
			Chroma:    89,
		},
		"cat_black.jpg": {
			Colors:    IndexedColors{0x2, 0x1, 0x1, 0x1, 0x2, 0x1, 0x2, 0x5, 0x2},
			MainColor: 1,
			Luminance: LightMap{0x8, 0xc, 0x9, 0x4, 0x2, 0x7, 0xd, 0xd, 0x3},
			Chroma:    9,
		},
		"cat_brown.jpg": {
			Colors:    IndexedColors{0x9, 0x5, 0x1, 0x2, 0x2, 0x1, 0x0, 0x6, 0x2},
			MainColor: 5,
			Luminance: LightMap{0x4, 0x5, 0xb, 0x4, 0x7, 0x3, 0x2, 0x5, 0x7},
			Chroma:    13,
		},
		"cat_yellow_grey.jpg": {
			Colors:    IndexedColors{0x2, 0x1, 0x1, 0x9, 0x0, 0x5, 0xb, 0x0, 0x5},
			MainColor: 5,
			Luminance: LightMap{0x9, 0x5, 0xb, 0x6, 0x1, 0x6, 0xa, 0x1, 0x8},
			Chroma:    20,
		},
	}

	err := filepath.Walk(conf.ExamplesPath(), func(filename string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if fileInfo.IsDir() || strings.HasPrefix(filepath.Base(filename), ".") {
			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !mediaFile.IsJpeg() {
			return nil
		}

		t.Run(filename, func(t *testing.T) {
			p, err := mediaFile.Colors(thumbsPath)

			basename := filepath.Base(filename)

			t.Log(p, err)

			assert.Nil(t, err)
			assert.True(t, p.Chroma.Int() >= 0)
			assert.True(t, p.Chroma.Int() <= 100)
			assert.NotEmpty(t, p.MainColor.Name())

			if e, ok := expected[basename]; ok {
				assert.Equal(t, e, p)
			}
		})

		return nil
	})

	if err != nil {
		t.Log(err.Error())
	}
}

func TestMediaFile_Colors(t *testing.T) {
	conf := config.TestConfig()

	t.Run("cat_brown.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/cat_brown.jpg"); err == nil {
			p, err := mediaFile.Colors(conf.ThumbnailsPath())

			t.Log(p, err)

			assert.Nil(t, err)
			assert.Equal(t, 13, p.Chroma.Int())
			assert.Equal(t, "D", p.Chroma.Hex())
			assert.IsType(t, IndexedColors{}, p.Colors)
			assert.Equal(t, "gold", p.MainColor.Name())
			assert.Equal(t, IndexedColors{0x9, 0x5, 0x1, 0x2, 0x2, 0x1, 0x0, 0x6, 0x2}, p.Colors)
			assert.Equal(t, LightMap{0x4, 0x5, 0xb, 0x4, 0x7, 0x3, 0x2, 0x5, 0x7}, p.Luminance)
		} else {
			t.Error(err)
		}
	})

	t.Run("fern_green.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/fern_green.jpg"); err == nil {
			p, err := mediaFile.Colors(conf.ThumbnailsPath())

			t.Log(p, err)

			assert.Nil(t, err)
			assert.Equal(t, 51, p.Chroma.Int())
			assert.Equal(t, "33", p.Chroma.Hex())
			assert.IsType(t, IndexedColors{}, p.Colors)
			assert.Equal(t, "lime", p.MainColor.Name())
			assert.Equal(t, IndexedColors{0xa, 0x9, 0xa, 0x9, 0xa, 0xa, 0x9, 0x9, 0x9}, p.Colors)
			assert.Equal(t, LightMap{0xb, 0x4, 0xa, 0x6, 0x9, 0x8, 0x2, 0x3, 0x4}, p.Luminance)
		} else {
			t.Error(err)
		}
	})

	t.Run("IMG_4120.JPG", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/IMG_4120.JPG"); err == nil {
			p, err := mediaFile.Colors(conf.ThumbnailsPath())

			t.Log(p, err)

			assert.Nil(t, err)
			assert.Equal(t, 7, p.Chroma.Int())
			assert.Equal(t, "7", p.Chroma.Hex())
			assert.IsType(t, IndexedColors{}, p.Colors)
			assert.Equal(t, "blue", p.MainColor.Name())
			assert.Equal(t, IndexedColors{0x2, 0x6, 0x6, 0x2, 0x2, 0x9, 0x2, 0x0, 0x0}, p.Colors)
		} else {
			t.Error(err)
		}
	})

	t.Run("leaves_gold.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/leaves_gold.jpg"); err == nil {
			p, err := mediaFile.Colors(conf.ThumbnailsPath())

			t.Log(p, err)

			assert.Nil(t, err)
			assert.Equal(t, 16, p.Chroma.Int())
			assert.Equal(t, "10", p.Chroma.Hex())
			assert.IsType(t, IndexedColors{}, p.Colors)
			assert.Equal(t, "gold", p.MainColor.Name())

			assert.Equal(t, IndexedColors{0x0, 0x0, 0x1, 0x5, 0x5, 0x0, 0x1, 0x5, 0x0}, p.Colors)
		} else {
			t.Error(err)
		}
	})
}
