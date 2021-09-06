package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbeddingsMidpoint(t *testing.T) {
	t.Run("2 embeddings, 1 dimension", func(t *testing.T) {
		e := Embeddings{Embedding{1}, Embedding{3}}

		result, r, c := EmbeddingsMidpoint(e)

		assert.Equal(t, Embedding{2}, result)
		assert.Equal(t, 1.01, r)
		assert.Equal(t, 2, c)
	})
	t.Run("3 embeddings, 1 dimension", func(t *testing.T) {
		e := Embeddings{Embedding{1}, Embedding{3}, Embedding{4}}

		result, r, c := EmbeddingsMidpoint(e)

		assert.Equal(t, Embedding{2.6666666666666665}, result)
		assert.Equal(t, 1.6766666666666665, r)
		assert.Equal(t, 3, c)
	})
	t.Run("4 embeddings, 1 dimension", func(t *testing.T) {
		e := Embeddings{Embedding{1}, Embedding{3}, Embedding{4}, Embedding{8}}

		result, r, c := EmbeddingsMidpoint(e)

		assert.Equal(t, Embedding{4}, result)
		assert.Equal(t, 4.01, r)
		assert.Equal(t, 4, c)
	})
	t.Run("empty embedding", func(t *testing.T) {
		e := Embeddings{}

		result, r, c := EmbeddingsMidpoint(e)

		assert.Len(t, result, 0)
		assert.Equal(t, float64(0), r)
		assert.Equal(t, 0, c)
	})
	t.Run("embedding with different length", func(t *testing.T) {
		e := Embeddings{Embedding{1}, Embedding{3, 5}}

		result, r, c := EmbeddingsMidpoint(e)

		assert.Equal(t, Embedding{2}, result)
		assert.Equal(t, 1.01, r)
		assert.Equal(t, 2, c)
	})
	t.Run("vectors", func(t *testing.T) {
		e := Embeddings{Embedding{1, 0, 0, 0}, Embedding{0, 1, 0, 0}, Embedding{0, 0, 1, 0}, Embedding{0, 0, 0, 1}}

		result, radius, count := EmbeddingsMidpoint(e)

		assert.Equal(t, Embedding{0.25, 0.25, 0.25, 0.25}, result)
		assert.Greater(t, 0.87612, radius)
		assert.Less(t, 0.8760, radius)
		assert.Equal(t, 4, count)
	})
}

func TestUnmarshalEmbedding(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r := UnmarshalEmbedding("[-0.013,-0.031]")
		assert.Equal(t, []float64{-0.013, -0.031}, r)
	})
	t.Run("no prefix", func(t *testing.T) {
		r := UnmarshalEmbedding("-0.013,-0.031]")
		assert.Nil(t, r)
	})
	t.Run("invalid json", func(t *testing.T) {
		r := UnmarshalEmbedding("[true, false]")
		assert.Equal(t, []float64{0, 0}, r)
	})
}

func TestUnmarshalEmbeddings(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r := UnmarshalEmbeddings("[[-0.013,-0.031]]")
		assert.Equal(t, [][]float64{{-0.013, -0.031}}, r)
	})
	t.Run("no prefix", func(t *testing.T) {
		r := UnmarshalEmbeddings("-0.013,-0.031]")
		assert.Nil(t, r)
	})
	t.Run("invalid json", func(t *testing.T) {
		r := UnmarshalEmbeddings("[[true, false]]")
		assert.Equal(t, [][]float64{{0, 0}}, r)
	})
}
