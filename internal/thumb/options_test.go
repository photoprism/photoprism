package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_Contains(t *testing.T) {
	t.Run("Left224", func(t *testing.T) {
		options := SizeLeft224.Options
		assert.True(t, options.Contains(ResampleFillTopLeft))
		assert.False(t, options.Contains(ResampleFillBottomRight))
		assert.False(t, options.Contains(ResampleFillCenter))
	})
	t.Run("Right224", func(t *testing.T) {
		options := SizeRight224.Options
		assert.False(t, options.Contains(ResampleFillTopLeft))
		assert.True(t, options.Contains(ResampleFillBottomRight))
		assert.False(t, options.Contains(ResampleFillCenter))
	})
	t.Run("Tile224", func(t *testing.T) {
		options := SizeTile224.Options
		assert.False(t, options.Contains(ResampleFillTopLeft))
		assert.False(t, options.Contains(ResampleFillBottomRight))
		assert.True(t, options.Contains(ResampleFillCenter))
	})
	t.Run("Tile500", func(t *testing.T) {
		options := SizeTile500.Options
		assert.False(t, options.Contains(ResampleFillTopLeft))
		assert.False(t, options.Contains(ResampleFillBottomRight))
		assert.True(t, options.Contains(ResampleFillCenter))
	})
	t.Run("Fit1600", func(t *testing.T) {
		options := SizeFit1600.Options
		assert.False(t, options.Contains(ResampleFillTopLeft))
		assert.False(t, options.Contains(ResampleFillBottomRight))
		assert.False(t, options.Contains(ResampleFillCenter))
	})
}
