package vector

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEuclideanDist(t *testing.T) {
	a := Vector{1, 2, 3, 4, 6, 5}
	b := Vector{2, 1, 3, 4, 5, 6}
	d := Vector{0, 0, 0, 0, 0, 0}
	e := Vector{}
	n := make(Vector, 512)

	t.Run("Method", func(t *testing.T) {
		assert.InDelta(t, 2, a.EuclideanDist(b), 0.01)
		assert.InDelta(t, a.EuclideanDist(b), b.EuclideanDist(a), 0.01)
		assert.True(t, math.IsNaN(faceEmbeddingB.EuclideanDist(d)))
		assert.InDelta(t, 0, d.EuclideanDist(d), 0.01)
		assert.True(t, math.IsNaN(e.EuclideanDist(d)))
		assert.InDelta(t, 0.9999999779072661, faceEmbeddingB.EuclideanDist(n), 0.01)
	})
	t.Run("Func", func(t *testing.T) {
		assert.InDelta(t, 2.0, EuclideanDist(a, b), 0.01)
	})
}

func TestCosineSimilarity(t *testing.T) {
	a := Vector{1, 2, 3, 4, 6, 5}
	b := Vector{2, 1, 3, 4, 5, 6}
	d := Vector{0, 0, 0, 0, 0, 0}
	e := Vector{}
	n := make(Vector, 512)

	t.Run("Values", func(t *testing.T) {
		assert.InDelta(t, 0.978021978021978, a.CosineSimilarity(b), 0.01)
		assert.True(t, math.IsNaN(faceEmbeddingB.CosineSimilarity(d)))
		assert.InDelta(t, 0, d.CosineSimilarity(d), 0.01)
		assert.True(t, math.IsNaN(e.CosineSimilarity(d)))
		assert.InDelta(t, 0, faceEmbeddingB.CosineSimilarity(n), 0.01)
		assert.InDelta(t, 0, n.CosineSimilarity(n), 0.01)
		assert.InDelta(t, 1.0, faceEmbeddingB.CosineSimilarity(faceEmbeddingB), 0.01)
		assert.InDelta(t, 1.0, a.CosineSimilarity(a), 0.01)
		assert.InDelta(t, 1.0, b.CosineSimilarity(b), 0.01)
	})
	t.Run("Func", func(t *testing.T) {
		assert.InDelta(t, 0.978021978021978, CosineSimilarity(a, b), 0.01)
	})
	t.Run("Orthogonal", func(t *testing.T) {
		assert.InDelta(t, 0.0, CosineSimilarity(Vector{1, 0}, Vector{0, 1}), 0.00001)
		assert.InDelta(t, 1.0, CosineDist(Vector{1, 0}, Vector{0, 1}), 0.00001)
	})
	t.Run("Opposite", func(t *testing.T) {
		assert.InDelta(t, -1.0, CosineSimilarity(Vector{1, 0}, Vector{-1, 0}), 0.00001)
		assert.InDelta(t, 2.0, CosineDist(Vector{1, 0}, Vector{-1, 0}), 0.00001)
	})
	t.Run("DimensionMismatch", func(t *testing.T) {
		assert.True(t, math.IsNaN(CosineSimilarity(Vector{1, 0}, Vector{1, 0, 0})))
		assert.True(t, math.IsNaN(CosineDist(Vector{1, 0}, Vector{1, 0, 0})))
	})
}

func TestCosineDist(t *testing.T) {
	a := Vector{1, 2, 3, 4, 6, 5}
	b := Vector{2, 1, 3, 4, 5, 6}
	d := Vector{0, 0, 0, 0, 0, 0}
	e := Vector{}
	n := make(Vector, 512)

	t.Run("Values", func(t *testing.T) {
		// Distance is 1 - similarity: 0 for identical, 1 for a zero vector, NaN on dim mismatch.
		assert.InDelta(t, 0.021978021978022, a.CosineDist(b), 0.01)
		assert.True(t, math.IsNaN(faceEmbeddingB.CosineDist(d)))
		assert.InDelta(t, 1.0, d.CosineDist(d), 0.01)
		assert.True(t, math.IsNaN(e.CosineDist(d)))
		assert.InDelta(t, 1.0, faceEmbeddingB.CosineDist(n), 0.01)
		assert.InDelta(t, 1.0, n.CosineDist(n), 0.01)
		assert.InDelta(t, 0, faceEmbeddingB.CosineDist(faceEmbeddingB), 0.01)
		assert.InDelta(t, 0, a.CosineDist(a), 0.01)
		assert.InDelta(t, 0, b.CosineDist(b), 0.01)
	})
	t.Run("Func", func(t *testing.T) {
		assert.InDelta(t, 0.021978021978022, CosineDist(a, b), 0.01)
	})
	t.Run("Equal", func(t *testing.T) {
		x := Vector{1, 0, 0, 1, 0, 0}
		y := Vector{1, 0, 0, 1, 0, 0}
		// Identical vectors: similarity 1, distance 0.
		assert.InDelta(t, 1.0, CosineSimilarity(x, y), 0.00001)
		assert.InDelta(t, 0.0, CosineDist(x, y), 0.00001)
	})
	t.Run("Faces", func(t *testing.T) {
		// Real face embeddings: near-orthogonal, so distance is ~1.
		assert.InDelta(t, -0.003275301858301365, CosineSimilarity(faceEmbeddingA, faceEmbeddingB), 0.00001)
		assert.InDelta(t, 1.003275301858301365, CosineDist(faceEmbeddingA, faceEmbeddingB), 0.00001)
	})
}

func TestCosineDists(t *testing.T) {
	x := Vectors{{1, 0}, {0, 1}}
	y := Vectors{{1, 0}, {-1, 0}}
	got := CosineDists(x, y)
	assert.Len(t, got, 2)
	assert.InDelta(t, 0.0, got[0][0], 0.00001) // identical
	assert.InDelta(t, 2.0, got[0][1], 0.00001) // opposite
	assert.InDelta(t, 1.0, got[1][0], 0.00001) // orthogonal
	assert.InDelta(t, 1.0, got[1][1], 0.00001) // orthogonal
}

func BenchmarkCosineDist(b *testing.B) {
	for b.Loop() {
		CosineDist(faceEmbeddingA, faceEmbeddingB)
	}
}
