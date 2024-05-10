package thumb

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
