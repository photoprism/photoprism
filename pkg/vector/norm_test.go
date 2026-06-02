package vector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVector_EuclideanNorm(t *testing.T) {
	a := Vector{1, 2, 3, 4, 6, 5}
	b := Vector{2, 1, 3, 4, 5, 6}
	d := Vector{0, 0, 0, 0, 0, 0}
	e := Vector{}

	assert.InDelta(t, 9.539392014169456, a.EuclideanNorm(), 0.01)
	assert.InDelta(t, 9.539392014169456, b.EuclideanNorm(), 0.01)
	assert.Equal(t, a.EuclideanNorm(), b.EuclideanNorm())
	assert.InDelta(t, 0.9999999779072661, faceEmbeddingB.EuclideanNorm(), 0.01)
	assert.InDelta(t, 0, d.EuclideanNorm(), 0.01)
	assert.InDelta(t, 0, e.EuclideanNorm(), 0.01)
	assert.Equal(t, d.EuclideanNorm(), e.EuclideanNorm())
}

func TestNorm(t *testing.T) {
	t.Run("Euclidean", func(t *testing.T) {
		assert.InDelta(t, 5.0, Norm(Vector{3, -4}, 2.0), 0.00001)
	})
	t.Run("Manhattan", func(t *testing.T) {
		// L1 norm uses absolute values, so negatives contribute their magnitude.
		assert.InDelta(t, 5.0, Norm(Vector{1, -2, 2}, 1.0), 0.00001)
	})
}
