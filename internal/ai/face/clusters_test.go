package face

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClustersContains(t *testing.T) {
	t.Run("InsideRadius", func(t *testing.T) {
		clusters := Clusters{
			{Radius: 0.5, Embedding: Embedding{0, 0}},
		}

		embedding := Embedding{0.2, 0.3}

		assert.True(t, clusters.Contains(embedding))
	})

	t.Run("OutsideRadius", func(t *testing.T) {
		clusters := Clusters{
			{Radius: 0.4, Embedding: Embedding{0, 0}},
		}

		embedding := Embedding{0.5, 0.5}

		assert.False(t, clusters.Contains(embedding))
	})

	t.Run("DisabledClusterBackground", func(t *testing.T) {
		clusters := Clusters{
			{Radius: 1, Embedding: Embedding{0, 0}, Disabled: true},
		}

		embedding := Embedding{0.1, 0.1}

		assert.False(t, clusters.Contains(embedding))
	})
}

func TestClustersDist(t *testing.T) {
	t.Run("ReturnsMinDistance", func(t *testing.T) {
		clusters := Clusters{
			{Radius: 0.4, Embedding: Embedding{0, 0}},
			{Radius: 0.4, Embedding: Embedding{1, 0}},
			{Radius: 0.4, Embedding: Embedding{0, 1}, Disabled: true},
		}

		embedding := Embedding{0.1, 0.2}
		dist := clusters.Dist(embedding)

		assert.InDelta(t, math.Sqrt(0.05), dist, 1e-9)
	})

	t.Run("NoEnabledClusters", func(t *testing.T) {
		clusters := Clusters{
			{Radius: 0.2, Embedding: Embedding{0, 0}, Disabled: true},
		}

		embedding := Embedding{0.1, 0.1}

		assert.Equal(t, float64(-1), clusters.Dist(embedding))
	})
}
