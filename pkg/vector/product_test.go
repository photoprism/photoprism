package vector

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProduct(t *testing.T) {
	t.Run("Values", func(t *testing.T) {
		p, err := Product(Vector{1, 2, 3}, Vector{4, 5, 6})
		assert.NoError(t, err)
		assert.Equal(t, Vector{4, 10, 18}, p)
	})
	t.Run("LengthMismatch", func(t *testing.T) {
		p, err := Product(Vector{1, 2, 3}, Vector{4, 5})
		assert.Error(t, err)
		assert.Nil(t, p)
	})
}

func TestDotProduct(t *testing.T) {
	t.Run("Values", func(t *testing.T) {
		r, err := DotProduct(Vector{1, 2, 3}, Vector{4, 5, 6})
		assert.NoError(t, err)
		assert.InDelta(t, 32.0, r, 0.00001)
	})
	t.Run("LengthMismatch", func(t *testing.T) {
		r, err := DotProduct(Vector{1, 2, 3}, Vector{4, 5})
		assert.Error(t, err)
		assert.True(t, math.IsNaN(r))
	})
}
