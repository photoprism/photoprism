package thumb

import (
	"errors"
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
	t.Run("ColorsBadIccProfile", func(t *testing.T) {
		// Regression for #5612 / #5616: Samsung Galaxy JPEGs ship an ICC
		// profile whose 4-byte size header is off by two bytes. SizeColors
		// declares ResampleStripICC, so the broken profile is dropped before
		// export and indexing succeeds.
		colorThumb := Sizes[Colors]
		src := "testdata/icc_profile_bad_length.jpg"
		dst := "testdata/vips/1/4/4/144456789098765432_3x3_resize.png"

		assert.FileExists(t, src)

		fileName, _, err := Vips(src, nil, "144456789098765432", "testdata/vips", colorThumb.Width, colorThumb.Height, colorThumb.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(fileName, dst))
		assert.FileExists(t, dst)

		dstimg, err := vips.LoadImageFromFile(dst, vips.NewImportParams())
		assert.NoError(t, err)
		defer dstimg.Close()
		assert.False(t, dstimg.HasICCProfile(), "SizeColors output must not embed an ICC profile")
	})
	t.Run("ColorsStripsValidIccProfile", func(t *testing.T) {
		// Confirms ResampleStripICC strips the profile even when libpng would
		// have accepted it: SizeColors is a temporary 3x3 cache consumed for
		// raw-pixel sampling, so an ICC chunk is wasted bytes on every photo.
		colorThumb := Sizes[Colors]
		src := "testdata/interop_index_srgb_icc.jpg"
		dst := "testdata/vips/1/5/5/155456789098765432_3x3_resize.png"

		assert.FileExists(t, src)

		srcimg, err := vips.LoadImageFromFile(src, vips.NewImportParams())
		assert.NoError(t, err)
		assert.True(t, srcimg.HasICCProfile(), "fixture sanity: input must carry an ICC profile")
		srcimg.Close()

		fileName, _, err := Vips(src, nil, "155456789098765432", "testdata/vips", colorThumb.Width, colorThumb.Height, colorThumb.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(fileName, dst))
		assert.FileExists(t, dst)

		dstimg, err := vips.LoadImageFromFile(dst, vips.NewImportParams())
		assert.NoError(t, err)
		defer dstimg.Close()
		assert.False(t, dstimg.HasICCProfile(), "ResampleStripICC must drop the profile")
	})
	t.Run("PngWithoutStripICCKeepsValidProfile", func(t *testing.T) {
		// Gates the new flag: a PNG export configured without ResampleStripICC
		// must preserve a valid ICC profile end-to-end. Guards against a
		// future regression that would generalize the strip to all PNGs.
		src := "testdata/interop_index_srgb_icc.jpg"
		dst := "testdata/vips/1/6/6/166456789098765432_3x3_resize.png"

		fileName, _, err := Vips(src, nil, "166456789098765432", "testdata/vips", 3, 3, ResampleResize, ResampleNearestNeighbor, ResamplePng)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(fileName, dst))
		assert.FileExists(t, dst)

		dstimg, err := vips.LoadImageFromFile(dst, vips.NewImportParams())
		assert.NoError(t, err)
		defer dstimg.Close()
		assert.True(t, dstimg.HasICCProfile(), "PNG without ResampleStripICC must preserve the profile")
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
		defer dstimg.Close()
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

func TestWrapVipsExportErr(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		inner := errors.New("vips2png: unable to write to target")
		err := wrapVipsExportErr("png", "/cache/1/2/3/colors.png", 3, 3, inner)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "png export failed")
		assert.Contains(t, err.Error(), "colors.png")
		assert.Contains(t, err.Error(), "3x3")
		assert.Contains(t, err.Error(), "unable to write to target")
		assert.True(t, errors.Is(err, inner), "wrapped error must remain unwrappable")
	})
	t.Run("StripsDirectory", func(t *testing.T) {
		// Only the basename should appear in the message so logs stay compact.
		err := wrapVipsExportErr("jpeg", "/cache/deep/nested/path/photo.jpg", 1920, 1080, errors.New("boom"))

		assert.Contains(t, err.Error(), "photo.jpg")
		assert.NotContains(t, err.Error(), "/cache/deep/nested")
	})
}

func TestWrapVipsWriteErr(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		inner := errors.New("no space left on device")
		err := wrapVipsWriteErr("/cache/1/2/3/colors.png", inner)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write thumbnail")
		assert.Contains(t, err.Error(), "colors.png")
		assert.Contains(t, err.Error(), "no space left on device")
		assert.True(t, errors.Is(err, inner), "wrapped error must remain unwrappable")
	})
	t.Run("StripsDirectory", func(t *testing.T) {
		// Only the basename should appear so the server filesystem layout is not leaked
		// to the UI via the event-bus log hook.
		err := wrapVipsWriteErr("/photoprism/storage/cache/thumbnails/1/2/3/photo.jpg", errors.New("boom"))

		assert.Contains(t, err.Error(), "photo.jpg")
		assert.NotContains(t, err.Error(), "/photoprism/storage")
	})
}
