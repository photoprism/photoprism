package face

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixtureEmbedding(t *testing.T) {
	t.Run("Deterministic", func(t *testing.T) {
		assert.Equal(t, FixtureEmbedding(7), FixtureEmbedding(7))
	})
	t.Run("UnitLength", func(t *testing.T) {
		e := FixtureEmbedding(1)
		require.Len(t, e, RandomEmbeddingDims())
		assert.InDelta(t, 0.0, e.Dist(e), 1e-9)
		assert.InDelta(t, 1.0, embeddingLength(e), 1e-9)
	})
	t.Run("SeedsAreFarApart", func(t *testing.T) {
		// Two seeds stand for two different people, so they have to land outside the widest
		// accept distance that can be configured. Independent unit vectors average sqrt(2)
		// apart, close enough to the 1.4 runtime cap that the configurable limit is the bar.
		for seed := uint64(2); seed < 8; seed++ {
			assert.Greater(t, FixtureEmbedding(1).Dist(FixtureEmbedding(seed)), float64(ConfigDistMax), seed)
		}
	})
	t.Run("FollowsConfiguredDims", func(t *testing.T) {
		assert.Len(t, FixtureEmbedding(3), ExpectedDims())
	})
}

func TestFixtureEmbeddingAt(t *testing.T) {
	base := FixtureEmbedding(1)

	t.Run("Distance", func(t *testing.T) {
		for _, dist := range []float64{0.05, 0.25, 0.5, 0.82, 1.2} {
			at := FixtureEmbeddingAt(base, dist, 42)
			assert.InDelta(t, dist, base.Dist(at), 1e-9, dist)
			assert.InDelta(t, 1.0, embeddingLength(at), 1e-9, dist)
		}
	})
	t.Run("Deterministic", func(t *testing.T) {
		assert.Equal(t, FixtureEmbeddingAt(base, 0.3, 5), FixtureEmbeddingAt(base, 0.3, 5))
		assert.NotEqual(t, FixtureEmbeddingAt(base, 0.3, 5), FixtureEmbeddingAt(base, 0.3, 6))
	})
	t.Run("ZeroDistance", func(t *testing.T) {
		assert.Equal(t, base, FixtureEmbeddingAt(base, 0, 5))
		assert.Equal(t, base, FixtureEmbeddingAt(base, -1, 5))
	})
	t.Run("EmptyBase", func(t *testing.T) {
		assert.Empty(t, FixtureEmbeddingAt(Embedding{}, 0.5, 5))
	})
	t.Run("DoesNotModifyBase", func(t *testing.T) {
		before := make(Embedding, len(base))
		copy(before, base)

		FixtureEmbeddingAt(base, 0.7, 9)

		assert.Equal(t, before, base)
	})
}

// embeddingLength returns the L2 norm of an embedding.
func embeddingLength(e Embedding) float64 {
	var sum float64

	for _, v := range e {
		sum += v * v
	}

	return math.Sqrt(sum)
}
