package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScaleScoreThreshold(t *testing.T) {
	t.Run("XXS", func(t *testing.T) {
		assert.Equal(t, float32(35), ScaleScoreThreshold(21))
	})
	t.Run("XS", func(t *testing.T) {
		assert.Equal(t, float32(25), ScaleScoreThreshold(27))
	})
	t.Run("S", func(t *testing.T) {
		assert.Equal(t, float32(20), ScaleScoreThreshold(33))
	})
	t.Run("M", func(t *testing.T) {
		assert.Equal(t, float32(18), ScaleScoreThreshold(45))
	})
	t.Run("L", func(t *testing.T) {
		assert.Equal(t, float32(15), ScaleScoreThreshold(75))
	})
	t.Run("XL", func(t *testing.T) {
		assert.Equal(t, float32(11), ScaleScoreThreshold(100))
	})
	t.Run("XXL", func(t *testing.T) {
		assert.Equal(t, float32(9), ScaleScoreThreshold(250))
	})
}
