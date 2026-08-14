package face

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbedding_Dist(t *testing.T) {
	t.Run("EqualLength", func(t *testing.T) {
		a := Embedding{0, 0, 0}
		b := Embedding{3, 4, 0}

		assert.InDelta(t, 5.0, a.Dist(b), 1e-9)
		assert.InDelta(t, 5.0, b.Dist(a), 1e-9)
	})
	t.Run("MismatchedLength", func(t *testing.T) {
		a := Embedding{0, 0}
		b := Embedding{1}

		assert.Equal(t, -1.0, a.Dist(b))
	})
}

func TestNewEmbedding_Normalized(t *testing.T) {
	raw := []float32{3, 4}
	result := NewEmbedding(raw)

	var sum float64
	for _, v := range result {
		sum += v * v
	}

	assert.InDelta(t, 1.0, math.Sqrt(sum), 1e-9)
}

var benchmarkEmbeddingDist float64

// benchmarkEmbeddingPair returns deterministic embeddings for Dist benchmarks.
func benchmarkEmbeddingPair(dim int) (Embedding, Embedding) {
	a := make(Embedding, dim)
	other := make(Embedding, dim)

	scale := float64(dim)

	for i := 0; i < dim; i++ {
		a[i] = float64(i+1) / scale
		other[i] = float64(dim-i) / scale
	}

	return a, other
}

func BenchmarkEmbeddingDist(b *testing.B) {
	for _, dim := range []int{128, 512, 1024} {
		b.Run(fmt.Sprintf("Dim%d", dim), func(b *testing.B) {
			a, other := benchmarkEmbeddingPair(dim)

			b.SetBytes(int64(dim * 16))
			b.ReportAllocs()

			for b.Loop() {
				benchmarkEmbeddingDist = a.Dist(other)
			}
		})
	}
}

func TestUnmarshalEmbedding(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		emb, err := UnmarshalEmbedding("[-0.013,-0.031]")

		assert.NoError(t, err)
		assert.Len(t, emb, 2)

		var norm float64
		for _, v := range emb {
			norm += v * v
		}

		assert.InDelta(t, 1.0, math.Sqrt(norm), 1e-9)
		assert.Less(t, emb[0], 0.0)
		assert.Less(t, emb[1], 0.0)
	})
	t.Run("NoPrefix", func(t *testing.T) {
		emb, err := UnmarshalEmbedding("-0.013,-0.031]")

		assert.Error(t, err)
		assert.Nil(t, emb)
	})
	t.Run("InvalidValues", func(t *testing.T) {
		emb, err := UnmarshalEmbedding("[true, false]")

		assert.Error(t, err)
		assert.Equal(t, Embedding{0, 0}, emb)
	})
}
