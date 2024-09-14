package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs/fastwalk"
	"github.com/photoprism/photoprism/pkg/media/colors"
)

func TestMediaFile_Colors_Testdata(t *testing.T) {
	conf := config.TestConfig()

	thumbsPath := os.TempDir() + "/TestMediaFile_Colors_Testdata"
	defer os.RemoveAll(thumbsPath)

	/*
		TODO: Add and compare other images in "testdata/"
	*/
	expected := map[string]colors.ColorPerception{
		"elephant_mono.jpg": {
			Colors:    colors.Colors{0x1, 0x1, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0},
			MainColor: 0,
			Luminance: colors.LightMap{0xa, 0x9, 0x0, 0x0, 0x6, 0x0, 0x0, 0x0, 0x0},
			Chroma:    0,
		},
		"sharks_blue.jpg": {
			Colors:    colors.Colors{0x6, 0x6, 0x6, 0x6, 0x6, 0x6, 0x5, 0x5, 0x6},
			MainColor: 6,
			Luminance: colors.LightMap{0x9, 0x7, 0x5, 0x4, 0x3, 0x4, 0x3, 0x3, 0x3},
			Chroma:    89,
		},
		"cat_black.jpg": {
			Colors:    colors.Colors{0x1, 0x2, 0x2, 0x2, 0x1, 0x2, 0x1, 0x3, 0x1},
			MainColor: 2,
			Luminance: colors.LightMap{0x8, 0xc, 0x9, 0x4, 0x2, 0x7, 0xd, 0xd, 0x3},
			Chroma:    9,
		},
		"cat_brown.jpg": {
			Colors:    colors.Colors{0x9, 0x3, 0x2, 0x1, 0x1, 0x2, 0x0, 0x6, 0x1},
			MainColor: 3,
			Luminance: colors.LightMap{0x4, 0x5, 0xb, 0x4, 0x7, 0x3, 0x2, 0x5, 0x7},
			Chroma:    13,
		},
		"cat_yellow_grey.jpg": {
			Colors:    colors.Colors{0x1, 0x2, 0x2, 0x9, 0x0, 0x3, 0xb, 0x0, 0x3},
			MainColor: 3,
			Luminance: colors.LightMap{0x9, 0x5, 0xb, 0x6, 0x1, 0x6, 0xa, 0x1, 0x8},
			Chroma:    20,
		},
		"Screenshot 2019-05-21 at 10.45.52.png": {
			Colors:    colors.Colors{4, 4, 4, 4, 4, 4, 4, 4, 4},
			MainColor: 4,
			Luminance: colors.LightMap{14, 14, 14, 15, 15, 15, 15, 15, 15},
			Chroma:    1,
		},
	}

	if err := fastwalk.Walk(conf.ExamplesPath(), func(filename string, info os.FileMode) error {
		if info.IsDir() || strings.HasPrefix(filepath.Base(filename), ".") {
			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil {
			t.Fatal(err)
		}

		if !mediaFile.IsPreviewImage() {
			t.Logf("not a jpeg: %s", filepath.Base(mediaFile.FileName()))
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
	}); err != nil {
		t.Fatal(err)
	}
}

func TestMediaFile_Colors(t *testing.T) {
	c := config.TestConfig()

	t.Run("cat_brown.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(c.ExamplesPath() + "/cat_brown.jpg"); err == nil {
			file, fileErr := mediaFile.Colors(c.ThumbCachePath())

			t.Log(file, fileErr)

			assert.Nil(t, fileErr)
			assert.Equal(t, 13, file.Chroma.Int())
			assert.Equal(t, "D", file.Chroma.Hex())
			assert.IsType(t, colors.Colors{}, file.Colors)
			assert.Equal(t, "gold", file.MainColor.Name())
			assert.Equal(t, colors.Colors{0x9, 0x3, 0x2, 0x1, 0x1, 0x2, 0x0, 0x6, 0x1}, file.Colors)
			assert.Equal(t, colors.LightMap{0x4, 0x5, 0xb, 0x4, 0x7, 0x3, 0x2, 0x5, 0x7}, file.Luminance)
		} else {
			t.Error(err)
		}
	})
	t.Run("FernRegular", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(c.ExamplesPath() + "/fern_green.jpg"); err == nil {
			file, fileErr := mediaFile.Colors(c.ThumbCachePath())

			t.Log(file, fileErr)

			assert.Nil(t, fileErr)
			assert.Equal(t, 51, file.Chroma.Int())
			assert.Equal(t, "33", file.Chroma.Hex())
			assert.IsType(t, colors.Colors{}, file.Colors)
			assert.Equal(t, "lime", file.MainColor.Name())
			assert.Equal(t, colors.Colors{0xa, 0x9, 0xa, 0x9, 0xa, 0xa, 0x9, 0x9, 0x9}, file.Colors)
			assert.Equal(t, colors.LightMap{0xb, 0x4, 0xa, 0x6, 0x9, 0x8, 0x2, 0x3, 0x4}, file.Luminance)
		} else {
			t.Error(err)
		}
	})
	t.Run("FernLarge", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(c.ExamplesPath() + "/fern_green.jpg"); err == nil {
			thumbLarge := thumb.SizeColors
			thumbLarge.Height = 16
			thumbLarge.Width = 16
			thumbLarge.Name = "color_large"

			thumb.Sizes[thumb.Colors] = thumbLarge

			file, fileErr := mediaFile.Colors(c.ThumbCachePath())

			thumb.Sizes[thumb.Colors] = thumb.SizeColors

			t.Log(file, fileErr)

			assert.Nil(t, fileErr)
			assert.Equal(t, 67, file.Chroma.Int())
			assert.Equal(t, "43", file.Chroma.Hex())
			assert.IsType(t, colors.Colors{}, file.Colors)
			assert.Equal(t, "lime", file.MainColor.Name())
			assert.Equal(t, colors.Colors{9, 10, 10, 10, 10, 10, 10, 10, 10}, file.Colors)
			assert.Equal(t, colors.LightMap{5, 9, 9, 9, 9, 10, 10, 9, 11}, file.Luminance)
		} else {
			t.Error(err)
		}
	})
	t.Run("IMG_4120.JPG", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(c.ExamplesPath() + "/IMG_4120.JPG"); err == nil {
			file, fileErr := mediaFile.Colors(c.ThumbCachePath())

			t.Log(file, fileErr)

			assert.Nil(t, fileErr)
			assert.Equal(t, 7, file.Chroma.Int())
			assert.Equal(t, "7", file.Chroma.Hex())
			assert.IsType(t, colors.Colors{}, file.Colors)
			assert.Equal(t, "blue", file.MainColor.Name())
			assert.Equal(t, colors.Colors{0x1, 0x6, 0x6, 0x1, 0x1, 0x9, 0x1, 0x0, 0x0}, file.Colors)
		} else {
			t.Error(err)
		}
	})
	t.Run("leaves_gold.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(c.ExamplesPath() + "/leaves_gold.jpg"); err == nil {
			file, fileErr := mediaFile.Colors(c.ThumbCachePath())

			t.Log(file, fileErr)

			assert.Nil(t, fileErr)
			assert.Equal(t, 16, file.Chroma.Int())
			assert.Equal(t, "10", file.Chroma.Hex())
			assert.IsType(t, colors.Colors{}, file.Colors)
			assert.Equal(t, "gold", file.MainColor.Name())

			assert.Equal(t, colors.Colors{0x0, 0x0, 0x2, 0x3, 0x3, 0x0, 0x2, 0x3, 0x0}, file.Colors)
		} else {
			t.Error(err)
		}
	})
	t.Run("Random.docx", func(t *testing.T) {
		file, fileErr := NewMediaFile(c.ExamplesPath() + "/Random.docx")
		p, fileErr := file.Colors(c.ThumbCachePath())
		assert.Error(t, fileErr, "no color information: not a JPEG file")
		t.Log(p)
	})
	t.Run("animated-earth.thm", func(t *testing.T) {
		file, fileErr := NewMediaFile(c.ExamplesPath() + "/animated-earth.thm")
		p, fileErr := file.Colors(c.ThumbCachePath())
		assert.Error(t, fileErr, "no color information: not a JPEG file")
		t.Log(p)
	})
}
