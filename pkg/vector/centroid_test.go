package vector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCentroid(t *testing.T) {
	t.Run("Mean", func(t *testing.T) {
		got := Centroid(Vectors{{1, 2}, {3, 4}})
		assert.Equal(t, Vector{2, 3}, got)
	})
	t.Run("Single", func(t *testing.T) {
		got := Centroid(Vectors{{2, 4, 6}})
		assert.Equal(t, Vector{2, 4, 6}, got)
	})
	t.Run("SkipsMismatchedDimensions", func(t *testing.T) {
		// The 3-D vector is ignored; the mean is taken over the two 2-D vectors.
		got := Centroid(Vectors{{1, 2}, {1, 2, 3}, {3, 4}})
		assert.Equal(t, Vector{2, 3}, got)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, Centroid(Vectors{}))
		assert.Nil(t, Centroid(nil))
	})
	t.Run("FirstVectorEmpty", func(t *testing.T) {
		assert.Nil(t, Centroid(Vectors{{}, {1, 2}}))
	})
	t.Run("Independent", func(t *testing.T) {
		a := Vector{1, 2}
		b := Vector{3, 4}
		got := Centroid(Vectors{a, b})
		got[0] = 100
		// Mutating the result must not change the inputs.
		assert.Equal(t, Vector{1, 2}, a)
		assert.Equal(t, Vector{3, 4}, b)
	})
}
