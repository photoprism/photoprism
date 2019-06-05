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
		"testdata/elephant_mono.jpg": {
			Colors:    IndexedColors{0x2, 0x2, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0},
			MainColor: 0,
			Luminance: LightMap{0xa, 0x9, 0x0, 0x0, 0x6, 0x0, 0x0, 0x0, 0x0},
			Chroma:    0,
		},
		"testdata/sharks_blue.jpg": {
			Colors:    IndexedColors{0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x4, 0x4, 0x6},
			MainColor: 6,
			Luminance: LightMap{0x9, 0x7, 0x5, 0x4, 0x3, 0x4, 0x3, 0x3, 0x3},
			Chroma:    13,
		},
		"testdata/cat_black.jpg": {
			Colors:    IndexedColors{0x2, 0x1, 0x1, 0x1, 0x2, 0x1, 0x2, 0x5, 0x2},
			MainColor: 1,
			Luminance: LightMap{0x8, 0xc, 0x9, 0x4, 0x2, 0x7, 0xd, 0xd, 0x3},
			Chroma:    1,
		},
		"testdata/cat_brown.jpg": {
			Colors:    IndexedColors{0x9, 0x5, 0x1, 0x2, 0x2, 0x1, 0x0, 0x6, 0x2},
			MainColor: 5,
			Luminance: LightMap{0x4, 0x5, 0xb, 0x4, 0x7, 0x3, 0x2, 0x5, 0x7},
			Chroma:    2,
		},
		"testdata/cat_yellow_grey.jpg": {
			Colors:    IndexedColors{0x2, 0x1, 0x1, 0x9, 0x0, 0x5, 0xb, 0x0, 0x5},
			MainColor: 5,
			Luminance: LightMap{0x9, 0x5, 0xb, 0x6, 0x1, 0x6, 0xa, 0x1, 0x8},
			Chroma:    3,
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

			t.Log(p, err)

			assert.Nil(t, err)
			assert.True(t, p.Chroma.Int() >= 0)
			assert.True(t, p.Chroma.Int() < 16)
			assert.NotEmpty(t, p.MainColor.Name())

			if e, ok := expected[filename]; ok {
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

	conf.InitializeTestData(t)

	t.Run("dog.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ImportPath() + "/dog.jpg"); err == nil {
			p, err := mediaFile.Colors(conf.ThumbnailsPath())

			t.Log(p, err)

			assert.Nil(t, err)
			assert.Equal(t, 2, p.Chroma.Int())
			assert.IsType(t, IndexedColors{}, p.Colors)
			assert.Equal(t, "brown", p.MainColor.Name())
			assert.Equal(t, IndexedColors{0x1, 0x3, 0x1, 0x2, 0xe, 0x0, 0x2, 0x2, 0x0}, p.Colors)
			assert.Equal(t, LightMap{0x4, 0xf, 0x8, 0xc, 0x4, 0x2, 0x4, 0x3, 0x0}, p.Luminance)
		} else {
			t.Error(err)
		}
	})

	t.Run("ape.jpeg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ImportPath() + "/ape.jpeg"); err == nil {
			p, err := mediaFile.Colors(conf.ThumbnailsPath())

			t.Log(p, err)

			assert.Nil(t, err)
			assert.Equal(t, 3, p.Chroma.Int())
			assert.IsType(t, IndexedColors{}, p.Colors)
			assert.Equal(t, "green", p.MainColor.Name())
			assert.Equal(t, IndexedColors{0x9, 0x9, 0x9, 0x8, 0x9, 0x2, 0x9, 0x2, 0x9}, p.Colors)
			assert.Equal(t, LightMap{0x6, 0xb, 0x9, 0x7, 0x9, 0x6, 0x8, 0xa, 0xe}, p.Luminance)
		} else {
			t.Error(err)
		}
	})

	t.Run("iphone/IMG_6788.JPG", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ImportPath() + "/iphone/IMG_6788.JPG"); err == nil {
			p, err := mediaFile.Colors(conf.ThumbnailsPath())

			t.Log(p, err)

			assert.Nil(t, err)
			assert.Equal(t, 3, p.Chroma.Int())
			assert.IsType(t, IndexedColors{}, p.Colors)
			assert.Equal(t, "blue", p.MainColor.Name())
			assert.Equal(t, IndexedColors{0x6, 0x6, 0x9, 0x1, 0x0, 0x2, 0x1, 0x1, 0x6}, p.Colors)
		} else {
			t.Error(err)
		}
	})

	t.Run("raw/20140717_154212_1EC48F8489.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ImportPath() + "/raw/20140717_154212_1EC48F8489.jpg"); err == nil {
			p, err := mediaFile.Colors(conf.ThumbnailsPath())

			t.Log(p, err)

			assert.Nil(t, err)
			assert.Equal(t, 2, p.Chroma.Int())
			assert.IsType(t, IndexedColors{}, p.Colors)
			assert.Equal(t, "green", p.MainColor.Name())

			assert.Equal(t, IndexedColors{0x3, 0x3, 0x6, 0x1, 0x2, 0x2, 0x9, 0x9, 0x9}, p.Colors)
		} else {
			t.Error(err)
		}
	})
}
