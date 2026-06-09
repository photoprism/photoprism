package vector

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVector_Variance(t *testing.T) {
	t.Run("Values", func(t *testing.T) {
		assert.InDelta(t, 32.0/7.0, Vector{2, 4, 4, 4, 5, 5, 7, 9}.Variance(), 0.00001)
	})
	t.Run("Single", func(t *testing.T) {
		assert.InDelta(t, 0.0, Vector{5}.Variance(), 0.00001)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.InDelta(t, 0.0, Vector{}.Variance(), 0.00001)
	})
}

func TestVector_Sd(t *testing.T) {
	assert.InDelta(t, math.Sqrt(32.0/7.0), Vector{2, 4, 4, 4, 5, 5, 7, 9}.Sd(), 0.00001)
	assert.InDelta(t, 0.0, Vector{5}.Sd(), 0.00001)
}

func TestCor(t *testing.T) {
	t.Run("PerfectPositive", func(t *testing.T) {
		r, err := Cor(Vector{1, 2, 3}, Vector{2, 4, 6})
		assert.NoError(t, err)
		assert.InDelta(t, 1.0, r, 0.00001)
	})
	t.Run("LengthMismatch", func(t *testing.T) {
		r, err := Cor(Vector{1, 2, 3}, Vector{1, 2})
		assert.Error(t, err)
		assert.True(t, math.IsNaN(r))
	})
}
