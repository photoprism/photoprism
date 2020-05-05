package thumb

import (
	"os"
	"testing"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestResampleOptions(t *testing.T) {
	method, filter, format := ResampleOptions(ResamplePng, ResampleFillCenter, ResampleDefault)

	assert.Equal(t, ResampleFillCenter, method)
	assert.Equal(t, imaging.Lanczos.Support, filter.Support)
	assert.Equal(t, fs.TypePng, format)
}

func TestResample(t *testing.T) {
	tile50 := Types["tile_50"]

	src := "testdata/example.jpg"

	assert.FileExists(t, src)

	img, err := imaging.Open(src, imaging.AutoOrientation(true))

	if err != nil {
		t.Fatal(err)
	}

	bounds := img.Bounds()

	assert.Equal(t, 750, bounds.Max.X)
	assert.Equal(t, 500, bounds.Max.Y)

	result := *Resample(&img, tile50.Width, tile50.Height, tile50.Options...)

	boundsNew := result.Bounds()

	assert.Equal(t, 50, boundsNew.Max.X)
	assert.Equal(t, 50, boundsNew.Max.Y)
}

func TestPostfix(t *testing.T) {
	tile50 := Types["tile_50"]

	result := Postfix(tile50.Width, tile50.Height, tile50.Options...)

	assert.Equal(t, "50x50_center.jpg", result)
}

func TestFilename(t *testing.T) {
	t.Run("colors", func(t *testing.T) {
		colorThumb := Types["colors"]

		result, err := Filename("123456789098765432", "testdata", colorThumb.Width, colorThumb.Height, colorThumb.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "testdata/1/2/3/123456789098765432_3x3_resize.png", result)
	})

	t.Run("fit_720", func(t *testing.T) {
		fit720 := Types["fit_720"]

		result, err := Filename("123456789098765432", "testdata", fit720.Width, fit720.Height, fit720.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "testdata/1/2/3/123456789098765432_720x720_fit.jpg", result)
	})
}

func TestFromFile(t *testing.T) {
	t.Run("colors", func(t *testing.T) {
		colorThumb := Types["colors"]
		src := "testdata/example.gif"
		dst := "testdata/1/2/3/123456789098765432_3x3_resize.png"

		assert.FileExists(t, src)

		fileName, err := FromFile(src, "123456789098765432", "testdata", colorThumb.Width, colorThumb.Height, colorThumb.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, dst, fileName)

		assert.FileExists(t, dst)
	})

	t.Run("missing file", func(t *testing.T) {
		colorThumb := Types["colors"]
		src := "testdata/example.xxx"

		assert.NoFileExists(t, src)

		fileName, err := FromFile(src, "193456789098765432", "testdata", colorThumb.Width, colorThumb.Height, colorThumb.Options...)

		assert.Equal(t, "", fileName)
		assert.Error(t, err)
	})
}

func TestFromCache(t *testing.T) {
	t.Run("missing thumb", func(t *testing.T) {
		tile50 := Types["tile_50"]
		src := "testdata/example.jpg"

		assert.FileExists(t, src)

		fileName, err := FromCache(src, "193456789098765432", "testdata", tile50.Width, tile50.Height, tile50.Options...)

		assert.Equal(t, "", fileName)

		if err != ErrThumbNotCached {
			t.Fatal("ErrThumbNotCached expected")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		tile50 := Types["tile_50"]
		src := "testdata/example.xxx"

		assert.NoFileExists(t, src)

		fileName, err := FromCache(src, "193456789098765432", "testdata", tile50.Width, tile50.Height, tile50.Options...)

		assert.Equal(t, "", fileName)
		assert.Error(t, err)
	})
}

func TestCreate(t *testing.T) {
	t.Run("tile_500", func(t *testing.T) {
		tile500 := Types["tile_500"]
		src := "testdata/example.jpg"
		dst := "testdata/example.tile_500.jpg"

		assert.FileExists(t, src)
		assert.NoFileExists(t, dst)

		img, err := imaging.Open(src, imaging.AutoOrientation(true))

		if err != nil {
			t.Fatal(err)
		}

		bounds := img.Bounds()

		assert.Equal(t, 750, bounds.Max.X)
		assert.Equal(t, 500, bounds.Max.Y)

		resized, err := Create(&img, dst, tile500.Width, tile500.Height, tile500.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, dst)

		if err := os.Remove(dst); err != nil {
			t.Fatal(err)
		}

		imgNew := *resized
		boundsNew := imgNew.Bounds()

		assert.Equal(t, 500, boundsNew.Max.X)
		assert.Equal(t, 500, boundsNew.Max.Y)
	})
}
