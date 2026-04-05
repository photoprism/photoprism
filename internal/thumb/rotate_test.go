package thumb

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

// newTestImage creates a small non-square NRGBA image for rotation tests.
// The top-left pixel is red so the caller can verify orientation.
func newTestImage(width, height int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	// Fill with white.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	// Mark top-left pixel red.
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})

	return img
}

func TestRotate(t *testing.T) {
	const w, h = 4, 3

	t.Run("Unspecified", func(t *testing.T) {
		result := Rotate(newTestImage(w, h), OrientationUnspecified)
		assert.Equal(t, w, result.Bounds().Dx())
		assert.Equal(t, h, result.Bounds().Dy())
	})
	t.Run("Normal", func(t *testing.T) {
		result := Rotate(newTestImage(w, h), OrientationNormal)
		assert.Equal(t, w, result.Bounds().Dx())
		assert.Equal(t, h, result.Bounds().Dy())
	})
	t.Run("FlipH", func(t *testing.T) {
		result := Rotate(newTestImage(w, h), OrientationFlipH)
		assert.Equal(t, w, result.Bounds().Dx())
		assert.Equal(t, h, result.Bounds().Dy())

		// Top-left pixel should now be at top-right.
		r, _, _, _ := result.At(w-1, 0).RGBA()
		assert.Equal(t, uint32(0xffff), r, "red pixel should be at top-right after horizontal flip")
	})
	t.Run("Rotate180", func(t *testing.T) {
		result := Rotate(newTestImage(w, h), OrientationRotate180)
		assert.Equal(t, w, result.Bounds().Dx())
		assert.Equal(t, h, result.Bounds().Dy())

		// Top-left pixel should now be at bottom-right.
		r, _, _, _ := result.At(w-1, h-1).RGBA()
		assert.Equal(t, uint32(0xffff), r, "red pixel should be at bottom-right after 180 rotation")
	})
	t.Run("FlipV", func(t *testing.T) {
		result := Rotate(newTestImage(w, h), OrientationFlipV)
		assert.Equal(t, w, result.Bounds().Dx())
		assert.Equal(t, h, result.Bounds().Dy())

		// Top-left pixel should now be at bottom-left.
		r, _, _, _ := result.At(0, h-1).RGBA()
		assert.Equal(t, uint32(0xffff), r, "red pixel should be at bottom-left after vertical flip")
	})
	t.Run("Rotate90", func(t *testing.T) {
		result := Rotate(newTestImage(w, h), OrientationRotate90)
		// 90 CCW swaps dimensions.
		assert.Equal(t, h, result.Bounds().Dx())
		assert.Equal(t, w, result.Bounds().Dy())

		// Top-left pixel should now be at bottom-left.
		r, _, _, _ := result.At(0, w-1).RGBA()
		assert.Equal(t, uint32(0xffff), r, "red pixel should be at bottom-left after 90 CCW rotation")
	})
	t.Run("Rotate270", func(t *testing.T) {
		result := Rotate(newTestImage(w, h), OrientationRotate270)
		// 270 CCW swaps dimensions.
		assert.Equal(t, h, result.Bounds().Dx())
		assert.Equal(t, w, result.Bounds().Dy())

		// Top-left pixel should now be at top-right.
		r, _, _, _ := result.At(h-1, 0).RGBA()
		assert.Equal(t, uint32(0xffff), r, "red pixel should be at top-right after 270 CCW rotation")
	})
	t.Run("Transpose", func(t *testing.T) {
		result := Rotate(newTestImage(w, h), OrientationTranspose)
		// Transpose swaps dimensions.
		assert.Equal(t, h, result.Bounds().Dx())
		assert.Equal(t, w, result.Bounds().Dy())

		// Transpose mirrors along top-left to bottom-right diagonal: (x,y) -> (y,x).
		r, _, _, _ := result.At(0, 0).RGBA()
		assert.Equal(t, uint32(0xffff), r, "red pixel should stay at top-left after transpose")
	})
	t.Run("Transverse", func(t *testing.T) {
		result := Rotate(newTestImage(w, h), OrientationTransverse)
		// Transverse swaps dimensions.
		assert.Equal(t, h, result.Bounds().Dx())
		assert.Equal(t, w, result.Bounds().Dy())

		// Transverse mirrors along top-right to bottom-left diagonal.
		r, _, _, _ := result.At(h-1, w-1).RGBA()
		assert.Equal(t, uint32(0xffff), r, "red pixel should be at bottom-right after transverse")
	})
	t.Run("InvalidOrientation", func(t *testing.T) {
		result := Rotate(newTestImage(w, h), 99)
		assert.Equal(t, w, result.Bounds().Dx())
		assert.Equal(t, h, result.Bounds().Dy())
	})
}
