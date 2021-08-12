package entity

import (
	"encoding/json"
	"strings"
)

type Embedding = []float64

// Embeddings represents marker face embeddings.
type Embeddings = []Embedding

// EmbeddingsMidpoint returns the embeddings vector midpoint.
func EmbeddingsMidpoint(m Embeddings) (result Embedding) {
	for i, emb := range m {
		if i == 0 {
			result = emb
			continue
		}

		for j, val := range result {
			result[j] = (val + emb[j]) / 2
		}
	}

	return result
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
