package thumb

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResampleWithFilter verifies explicit interpolation does not change target geometry.
func TestResampleWithFilter(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 8, 4))
	source.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})

	t.Run("Resize", func(t *testing.T) {
		result := ResampleWithFilter(source, 4, 4, ResampleResize, ResampleCubic)
		assert.Equal(t, image.Rect(0, 0, 4, 4), result.Bounds())
	})
	t.Run("Fill", func(t *testing.T) {
		result := ResampleWithFilter(source, 4, 4, ResampleFillCenter, ResampleLinear)
		assert.Equal(t, image.Rect(0, 0, 4, 4), result.Bounds())
	})
}
