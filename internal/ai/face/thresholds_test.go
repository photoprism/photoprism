package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQualityThreshold(t *testing.T) {
	t.Run("XXS", func(t *testing.T) {
		assert.Equal(t, float32(35), QualityThreshold(21))
	})
	t.Run("XS", func(t *testing.T) {
		assert.Equal(t, float32(25), QualityThreshold(27))
	})
	t.Run("S", func(t *testing.T) {
		assert.Equal(t, float32(20), QualityThreshold(33))
	})
	t.Run("M", func(t *testing.T) {
		assert.Equal(t, float32(18), QualityThreshold(45))
	})
	t.Run("L", func(t *testing.T) {
		assert.Equal(t, float32(15), QualityThreshold(75))
	})
	t.Run("XL", func(t *testing.T) {
		assert.Equal(t, float32(11), QualityThreshold(100))
	})
	t.Run("XXL", func(t *testing.T) {
		assert.Equal(t, float32(9), QualityThreshold(250))
	})
}
