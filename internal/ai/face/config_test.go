package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPigoQualityThreshold(t *testing.T) {
	t.Run("XXS", func(t *testing.T) {
		assert.Equal(t, float32(21), PigoQualityThreshold(21))
	})
	t.Run("XS", func(t *testing.T) {
		assert.Equal(t, float32(17), PigoQualityThreshold(27))
	})
	t.Run("S", func(t *testing.T) {
		assert.Equal(t, float32(15), PigoQualityThreshold(33))
	})
	t.Run("M", func(t *testing.T) {
		assert.Equal(t, float32(13), PigoQualityThreshold(45))
	})
	t.Run("L", func(t *testing.T) {
		assert.Equal(t, float32(11), PigoQualityThreshold(75))
	})
	t.Run("XL", func(t *testing.T) {
		assert.Equal(t, float32(10), PigoQualityThreshold(100))
	})
	t.Run("XXL", func(t *testing.T) {
		assert.Equal(t, float32(9), PigoQualityThreshold(250))
	})
}
