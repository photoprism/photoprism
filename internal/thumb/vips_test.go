package thumb

import (
	"strings"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
)

func TestVips(t *testing.T) {
	t.Run("Colors", func(t *testing.T) {
		colorThumb := Sizes[Colors]
		src := "testdata/example.gif"
		dst := "testdata/vips/1/2/3/123456789098765432_3x3_resize.png"

		assert.FileExists(t, src)

		fileName, _, err := Vips(src, nil, "123456789098765432", "testdata/vips", colorThumb.Width, colorThumb.Height, colorThumb.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(fileName, dst))
		assert.FileExists(t, dst)
	})
	t.Run("InteropIndexColors", func(t *testing.T) {
		thumb := Sizes[Tile500]
		src := "testdata/interop_index.jpg"
		dst := "testdata/vips/1/3/3/133456789098765432_500x500_center.jpg"

		assert.FileExists(t, src)

		fileName, _, err := Vips(src, nil, "133456789098765432", "testdata/vips", thumb.Width, thumb.Height, thumb.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(fileName, dst))
		assert.Equal(t, fileName, dst)
		assert.FileExists(t, dst)

		dstimg, err := vips.LoadImageFromFile(dst, vips.NewImportParams())
		assert.NoError(t, err)
		assert.True(t, dstimg.HasICCProfile())
		assert.True(t, dstimg.IsColorSpaceSupported())
	})
	t.Run("Left224", func(t *testing.T) {
		thumb := SizeLeft224
		src := "testdata/fixed.jpg"
		dst := "testdata/vips/1/2/3/123456789098765432_224x224_left.jpg"

		assert.FileExists(t, src)

		fileName, _, err := Vips(src, nil, "123456789098765432", "testdata/vips", thumb.Width, thumb.Height, thumb.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(fileName, dst))
		assert.FileExists(t, dst)
	})
	t.Run("TwoTiles", func(t *testing.T) {
		large := Sizes[Tile500]
		small := Sizes[Tile224]
		srcName := "testdata/example.jpg"
		dstLarge := "testdata/vips/1/2/3/123456789098765432_500x500_center.jpg"
		dstSmall := "testdata/vips/1/2/3/123456789098765432_224x224_center.jpg"

		assert.FileExists(t, srcName)

		thumbName, thumbBuffer, err := Vips(srcName, nil, "123456789098765432", "testdata/vips", large.Width, large.Height, large.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(thumbName, dstLarge))
		assert.FileExists(t, dstLarge)

		thumbName, _, err = Vips(srcName, thumbBuffer, "123456789098765432", "testdata/vips", small.Width, small.Height, small.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(thumbName, dstSmall))
		assert.FileExists(t, dstSmall)
	})
	/* t.Run("Rotate", func(t *testing.T) {
		thumb := Sizes[Fit1920]
		src := "testdata/exif-6.jpg"
		dst := "testdata/rotate/1/2/3/123456789098765432_1920x1200_fit.jpg"

		assert.FileExists(t, src)

		fileName, _, err := Vips(src, "123456789098765432", "testdata/rotate", thumb.Width, thumb.Height, 0, thumb.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(fileName, dst))
		assert.FileExists(t, dst)
	}) */
	t.Run("Fit1920", func(t *testing.T) {
		thumb := Sizes[Fit1920]
		src := "testdata/example.jpg"
		dst := "testdata/vips/1/2/3/123456789098765432_1920x1200_fit.jpg"

		assert.FileExists(t, src)

		fileName, _, err := Vips(src, nil, "123456789098765432", "testdata/vips", thumb.Width, thumb.Height, thumb.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(fileName, dst))
		assert.FileExists(t, dst)
	})
	t.Run("FileNotFound", func(t *testing.T) {
		colorThumb := Sizes[Colors]
		src := "testdata/example.xxx"

		assert.NoFileExists(t, src)

		fileName, _, err := Vips(src, nil, "193456789098765432", "testdata/vips", colorThumb.Width, colorThumb.Height, colorThumb.Options...)

		assert.Equal(t, "", fileName)
		assert.Error(t, err)
	})
	t.Run("EmptyFilename", func(t *testing.T) {
		colorThumb := Sizes[Colors]

		fileName, _, err := Vips("", nil, "193456789098765432", "testdata/vips", colorThumb.Width, colorThumb.Height, colorThumb.Options...)

		if err == nil {
			t.Fatal("error expected")
		}
		assert.Equal(t, "", fileName)
		assert.Equal(t, "thumb: invalid file name ''", err.Error())
	})
}

func TestVipsImportParams(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		result := VipsImportParams()

		if result == nil {
			t.Fatal("result is nil")
		}

		assert.True(t, result.AutoRotate.Get())
		assert.False(t, result.FailOnError.Get())
	})
}

func TestVipsPngExportParams(t *testing.T) {
	t.Run("Standard", func(t *testing.T) {
		result := VipsPngExportParams(500, 500)

		if result == nil {
			t.Fatal("result is nil")
		}

		assert.False(t, result.Interlace)
		assert.Equal(t, vips.PngFilterNone, result.Filter)
		assert.Equal(t, 0, result.Quality)
		assert.Equal(t, 6, result.Compression)
	})
	t.Run("Small", func(t *testing.T) {
		result := VipsPngExportParams(3, 3)

		if result == nil {
			t.Fatal("result is nil")
		}

		assert.False(t, result.Interlace)
		assert.Equal(t, vips.PngFilterNone, result.Filter)
		assert.Equal(t, 0, result.Quality)
		assert.Equal(t, 0, result.Compression)
	})
}

func TestVipsJpegExportParams(t *testing.T) {
	t.Run("Standard", func(t *testing.T) {
		result := VipsJpegExportParams(1920, 1200)

		if result == nil {
			t.Fatal("result is nil")
		}

		assert.True(t, result.Interlace)
		assert.False(t, result.TrellisQuant)
		assert.False(t, result.OptimizeScans)
		assert.True(t, result.OptimizeCoding)
		assert.False(t, result.OvershootDeringing)
		assert.Equal(t, JpegQualityDefault.Int(), result.Quality)
	})
	t.Run("Small", func(t *testing.T) {
		result := VipsJpegExportParams(50, 50)

		if result == nil {
			t.Fatal("result is nil")
		}

		assert.True(t, result.Interlace)
		assert.False(t, result.TrellisQuant)
		assert.False(t, result.OptimizeScans)
		assert.False(t, result.OptimizeCoding)
		assert.False(t, result.OvershootDeringing)
		assert.Equal(t, JpegQualitySmall().Int(), result.Quality)
	})
}
