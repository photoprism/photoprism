package thumb

import (
	"os"
	"strings"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestVips(t *testing.T) {
	t.Run("Colors", func(t *testing.T) {
		colorThumb := Sizes[Colors]
		src := "testdata/example.gif"
		dst := "testdata/vips/1/2/3/123456789098765432_3x3_resize.png"

		assert.FileExists(t, src)

		fileName, err := Vips(src, "123456789098765432", "testdata/vips", colorThumb.Width, colorThumb.Height, OrientationNormal, colorThumb.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(fileName, dst))
		assert.FileExists(t, dst)
	})
	t.Run("Tile224", func(t *testing.T) {
		thumb := Sizes[Tile224]
		src := "testdata/fixed.jpg"
		dst := "testdata/vips/1/2/3/123456789098765432_224x224_center.jpg"

		assert.FileExists(t, src)

		fileName, err := Vips(src, "123456789098765432", "testdata/vips", thumb.Width, thumb.Height, 0, thumb.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(fileName, dst))
		assert.FileExists(t, dst)
	})
	t.Run("Tile500WithOrientation", func(t *testing.T) {
		thumb := Sizes[Tile500]
		src := "testdata/example.jpg"
		dst := "testdata/vips/1/2/3/123456789098765432_500x500_center.jpg"

		assert.FileExists(t, src)

		fileName, err := Vips(src, "123456789098765432", "testdata/vips", thumb.Width, thumb.Height, 3, thumb.Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(fileName, dst))
		assert.FileExists(t, dst)
	})
	/* t.Run("Rotate", func(t *testing.T) {
		thumb := Sizes[Fit1920]
		src := "testdata/exif-6.jpg"
		dst := "testdata/rotate/1/2/3/123456789098765432_1920x1200_fit.jpg"

		assert.FileExists(t, src)

		fileName, err := Vips(src, "123456789098765432", "testdata/rotate", thumb.Width, thumb.Height, 0, thumb.Options...)

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

		fileName, err := Vips(src, "123456789098765432", "testdata/vips", thumb.Width, thumb.Height, 0, thumb.Options...)

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

		fileName, err := Vips(src, "193456789098765432", "testdata/vips", colorThumb.Width, colorThumb.Height, OrientationNormal, colorThumb.Options...)

		assert.Equal(t, "", fileName)
		assert.Error(t, err)
	})
	t.Run("EmptyFilename", func(t *testing.T) {
		colorThumb := Sizes[Colors]

		fileName, err := Vips("", "193456789098765432", "testdata/vips", colorThumb.Width, colorThumb.Height, OrientationNormal, colorThumb.Options...)

		if err == nil {
			t.Fatal("error expected")
		}
		assert.Equal(t, "", fileName)
		assert.Equal(t, "thumb: invalid file name ''", err.Error())
	})
}

func TestVipsRotate(t *testing.T) {
	if err := os.MkdirAll("testdata/vips/rotate", fs.ModeDir); err != nil {
		t.Fatal(err)
	}
	t.Run("OrientationNormal", func(t *testing.T) {
		src := "testdata/example.jpg"
		dst := "testdata/vips/rotate/0.jpg"

		assert.FileExists(t, src)

		// Load image from file.
		img, err := vips.NewImageFromFile(src)

		if err != nil {
			t.Fatal(err)
		}

		if err = VipsRotate(img, OrientationNormal); err != nil {
			t.Fatal(err)
		}

		params := vips.NewJpegExportParams()
		imageBytes, _, exportErr := img.ExportJpeg(params)

		if exportErr != nil {
			t.Fatal(exportErr)
		}

		// Write thumbnail to file.
		if err = os.WriteFile(dst, imageBytes, fs.ModeFile); err != nil {
			t.Fatal(exportErr)
		}

		assert.FileExists(t, dst)
	})
	t.Run("OrientationRotate90", func(t *testing.T) {
		src := "testdata/example.jpg"
		dst := "testdata/vips/rotate/90.jpg"

		assert.FileExists(t, src)

		// Load image from file.
		img, err := vips.NewImageFromFile(src)

		if err != nil {
			t.Fatal(err)
		}

		if err = VipsRotate(img, OrientationRotate90); err != nil {
			t.Fatal(err)
		}

		params := vips.NewJpegExportParams()
		imageBytes, _, exportErr := img.ExportJpeg(params)

		if exportErr != nil {
			t.Fatal(exportErr)
		}

		// Write thumbnail to file.
		if err = os.WriteFile(dst, imageBytes, fs.ModeFile); err != nil {
			t.Fatal(exportErr)
		}

		assert.FileExists(t, dst)
	})
	t.Run("OrientationRotate180", func(t *testing.T) {
		src := "testdata/example.jpg"
		dst := "testdata/vips/rotate/180.jpg"

		assert.FileExists(t, src)

		// Load image from file.
		img, err := vips.NewImageFromFile(src)

		if err != nil {
			t.Fatal(err)
		}

		if err = VipsRotate(img, OrientationRotate180); err != nil {
			t.Fatal(err)
		}

		params := vips.NewJpegExportParams()
		imageBytes, _, exportErr := img.ExportJpeg(params)

		if exportErr != nil {
			t.Fatal(exportErr)
		}

		// Write thumbnail to file.
		if err = os.WriteFile(dst, imageBytes, fs.ModeFile); err != nil {
			t.Fatal(exportErr)
		}

		assert.FileExists(t, dst)
	})
	t.Run("OrientationRotate270", func(t *testing.T) {
		src := "testdata/example.jpg"
		dst := "testdata/vips/rotate/270.jpg"

		assert.FileExists(t, src)

		// Load image from file.
		img, err := vips.NewImageFromFile(src)

		if err != nil {
			t.Fatal(err)
		}

		if err = VipsRotate(img, OrientationRotate270); err != nil {
			t.Fatal(err)
		}

		params := vips.NewJpegExportParams()
		imageBytes, _, exportErr := img.ExportJpeg(params)

		if exportErr != nil {
			t.Fatal(exportErr)
		}

		// Write thumbnail to file.
		if err = os.WriteFile(dst, imageBytes, fs.ModeFile); err != nil {
			t.Fatal(exportErr)
		}

		assert.FileExists(t, dst)
	})
}
