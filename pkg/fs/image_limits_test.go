package fs

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

func TestExceedsPixelBudget(t *testing.T) {
	max := MaxImagePixels
	defer func() { MaxImagePixels = max }()
	t.Run("WithinBudget", func(t *testing.T) {
		MaxImagePixels = 1000
		assert.NoError(t, ExceedsPixelBudget(10, 10, 1))
	})
	t.Run("AtBudget", func(t *testing.T) {
		MaxImagePixels = 100
		assert.NoError(t, ExceedsPixelBudget(10, 10, 1))
	})
	t.Run("AboveBudget", func(t *testing.T) {
		MaxImagePixels = 99
		assert.True(t, errors.Is(ExceedsPixelBudget(10, 10, 1), ErrImageTooLarge))
	})
	t.Run("FramesCount", func(t *testing.T) {
		MaxImagePixels = 150
		assert.NoError(t, ExceedsPixelBudget(10, 10, 1))
		assert.True(t, errors.Is(ExceedsPixelBudget(10, 10, 2), ErrImageTooLarge))
	})
	t.Run("MissingFrameCount", func(t *testing.T) {
		MaxImagePixels = 99
		// A format that reports no frame count must still be bounded by its own geometry.
		assert.True(t, errors.Is(ExceedsPixelBudget(10, 10, 0), ErrImageTooLarge))
	})
	t.Run("Disabled", func(t *testing.T) {
		MaxImagePixels = 0
		assert.NoError(t, ExceedsPixelBudget(100000, 100000, 10))
	})
	t.Run("UnknownGeometry", func(t *testing.T) {
		MaxImagePixels = 1
		assert.NoError(t, ExceedsPixelBudget(0, 0, 1))
		assert.NoError(t, ExceedsPixelBudget(-4, 8, 1))
	})
	t.Run("NoOverflow", func(t *testing.T) {
		MaxImagePixels = 150000000
		// Two large dimensions must be counted in 64-bit rather than wrapping into a small
		// positive number that passes the check.
		assert.True(t, errors.Is(ExceedsPixelBudget(65536, 65536, 1), ErrImageTooLarge))
		// Three factors exceed what an int64 holds, so the frame count has to be applied
		// without completing the multiplication.
		assert.True(t, errors.Is(ExceedsPixelBudget(10000000, 10000000, 2000000), ErrImageTooLarge))
		assert.True(t, errors.Is(ExceedsPixelBudget(3037000500, 3037000500, 4), ErrImageTooLarge))
	})
}

func TestDecodeImageFile_PixelBudget(t *testing.T) {
	max := MaxImagePixels
	defer func() { MaxImagePixels = max }()
	t.Run("WithinBudget", func(t *testing.T) {
		MaxImagePixels = max
		img, name, err := DecodeImageFile("testdata/test.jpg")
		assert.NoError(t, err)
		assert.Equal(t, "jpeg", name)
		assert.NotNil(t, img)
	})
	t.Run("AboveBudget", func(t *testing.T) {
		MaxImagePixels = 4
		img, _, err := DecodeImageFile("testdata/test.jpg")
		assert.True(t, errors.Is(err, ErrImageTooLarge))
		assert.Nil(t, img)
	})
	t.Run("AboveBudgetData", func(t *testing.T) {
		data, err := os.ReadFile("testdata/test.jpg")
		assert.NoError(t, err)
		MaxImagePixels = 4
		img, _, err := DecodeImageData(data)
		assert.True(t, errors.Is(err, ErrImageTooLarge))
		assert.Nil(t, img)
	})
	t.Run("ConfigStaysAvailable", func(t *testing.T) {
		MaxImagePixels = 4
		// Reading the geometry must keep working, as callers use it to report why a file
		// was skipped.
		cfg, _, err := DecodeImageConfigFile("testdata/test.jpg")
		assert.NoError(t, err)
		assert.Greater(t, cfg.Width, 0)
	})
}

// budgetTestImage returns a small deterministic image for the format round-trip below.
func budgetTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 7, 5))

	for y := 0; y < 5; y++ {
		for x := 0; x < 7; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 30), G: uint8(y * 40), B: 128, A: 255})
		}
	}

	return img
}

// TestDecodeImageData_EveryFormat pins that reading the geometry before the pixel data
// leaves every decoder with a reader it can still consume, and that each of them is bounded.
func TestDecodeImageData_EveryFormat(t *testing.T) {
	max := MaxImagePixels
	defer func() { MaxImagePixels = max }()

	src := budgetTestImage()

	encoders := map[string]func(*bytes.Buffer) error{
		"jpeg": func(b *bytes.Buffer) error { return jpeg.Encode(b, src, nil) },
		"png":  func(b *bytes.Buffer) error { return png.Encode(b, src) },
		"gif":  func(b *bytes.Buffer) error { return gif.Encode(b, src, nil) },
		"bmp":  func(b *bytes.Buffer) error { return bmp.Encode(b, src) },
		"tiff": func(b *bytes.Buffer) error { return tiff.Encode(b, src, nil) },
	}

	for name, encode := range encoders {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, encode(&buf))
			data := buf.Bytes()

			MaxImagePixels = max
			img, format, err := DecodeImageData(data)
			require.NoError(t, err)
			assert.Equal(t, name, format)
			require.NotNil(t, img)
			assert.Equal(t, src.Bounds(), img.Bounds())

			MaxImagePixels = 4
			img, _, err = DecodeImageData(data)
			assert.ErrorIs(t, err, ErrImageTooLarge)
			assert.Nil(t, img)
		})
	}
}

func TestDecodeConfig(t *testing.T) {
	t.Run("Unsupported", func(t *testing.T) {
		cfg, name, err := decodeConfig(io.NewSectionReader(bytes.NewReader([]byte("nope")), 0, 4), imageFormatUnknown)
		assert.ErrorIs(t, err, errUnsupportedImageFormat)
		assert.Empty(t, name)
		assert.Equal(t, image.Config{}, cfg)
	})
	t.Run("Jpeg", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, jpeg.Encode(&buf, budgetTestImage(), nil))
		data := buf.Bytes()
		cfg, name, err := decodeConfig(io.NewSectionReader(bytes.NewReader(data), 0, int64(len(data))), imageFormatJPEG)
		assert.NoError(t, err)
		assert.Equal(t, "jpeg", name)
		assert.Equal(t, 7, cfg.Width)
		assert.Equal(t, 5, cfg.Height)
	})
}
