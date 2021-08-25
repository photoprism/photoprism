package entity

import (
	"encoding/json"
	"strings"

	"github.com/photoprism/photoprism/pkg/clusters"
)

type Embedding = []float64

// Embeddings represents an embedding cluster.
type Embeddings = []Embedding

// EmbeddingsMidpoint returns the embeddings vector midpoint.
func EmbeddingsMidpoint(m Embeddings) (result Embedding, radius float64, count int) {
	count = len(m)

	for i, emb := range m {
		if i == 0 {
			result = emb
			continue
		}

		if len(m[i]) != len(m[i-1]) {
			continue
		}

		for j, val := range result {
			result[j] = (val + emb[j]) / 2
		}

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
