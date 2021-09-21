package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScaleScoreThreshold(t *testing.T) {
	t.Run("Small", func(t *testing.T) {
		assert.Equal(t, float32(30), ScaleScoreThreshold(10))
	})
	t.Run("Medium", func(t *testing.T) {
		assert.Equal(t, float32(15), ScaleScoreThreshold(75))
	})
	t.Run("Large", func(t *testing.T) {
		assert.Equal(t, float32(9), ScaleScoreThreshold(200))
	})
}
