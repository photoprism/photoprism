package entity

import (
	"encoding/json"
	"strings"

	"github.com/montanaflynn/stats"
	"github.com/photoprism/photoprism/pkg/clusters"
)

type Embedding = []float64

// Embeddings represents an embedding cluster.
type Embeddings = []Embedding

// EmbeddingsMidpoint returns the embeddings vector midpoint.
func EmbeddingsMidpoint(embeddings Embeddings) (result Embedding, radius float64, count int) {
	count = len(embeddings)

	// No embeddings?
	if count == 0 {
		return result, radius, count
	} else if count == 1 {
		return embeddings[0], 0.0, count
	}

	dim := len(embeddings[0])

	// No embedding values?
	if dim == 0 {
		return Embedding{}, 0.0, count
	}

	result = make(Embedding, dim)

	// The mean of a set of vectors is calculated component-wise.
	for i := 0; i < dim; i++ {
		values := make(stats.Float64Data, count)

		for j := 0; j < count; j++ {
			values[j] = embeddings[j][i]
		}

		if m, err := stats.Mean(values); err != nil {
			log.Warnf("embeddings: %s", err)
		} else {
			result[i] = m
		}
	}

	// Radius is the max embedding distance + 0.01 from result.
	for _, emb := range embeddings {
		if d := clusters.EuclideanDistance(result, emb); d > radius {
			radius = d + 0.01
		}
	}

	return result, radius, count
}

// UnmarshalEmbeddings parses face embedding JSON.
func UnmarshalEmbeddings(s string) (result Embeddings) {
	if !strings.HasPrefix(s, "[[") {
		return nil
	}

	if err := json.Unmarshal([]byte(s), &result); err != nil {
		log.Errorf("faces: %s", err)
	}

	return result
}

// UnmarshalEmbedding parses a single face embedding JSON.
func UnmarshalEmbedding(s string) (result Embedding) {
	if !strings.HasPrefix(s, "[") {
		return nil
	}

	if err := json.Unmarshal([]byte(s), &result); err != nil {
		log.Errorf("faces: %s", err)
	}

	return result
}
