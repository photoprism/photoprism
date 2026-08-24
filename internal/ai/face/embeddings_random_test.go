package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomDist(t *testing.T) {
	t.Run("Range", func(t *testing.T) {
		d := RandomDist()
		assert.GreaterOrEqual(t, d, 0.1)
		assert.LessOrEqual(t, d, 1.5)
	})
}

func TestRandomEmbeddings(t *testing.T) {
	t.Run("Count", func(t *testing.T) {
		e := RandomEmbeddings(3, RegularFace)
		assert.Len(t, e, 3)

		for i := range e {
			assert.Len(t, e[i], RandomEmbeddingDims())
		}
	})
	t.Run("None", func(t *testing.T) {
		assert.Empty(t, RandomEmbeddings(0, RegularFace))
	})
	t.Run("Normalized", func(t *testing.T) {
		// Distances only mean anything on unit vectors, so fixtures have to be normalized too.
		for _, e := range RandomEmbeddings(2, RegularFace) {
			assert.InDelta(t, 1.0, e.Magnitude(), 0.001)
		}
	})
}
