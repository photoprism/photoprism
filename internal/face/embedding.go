package face

import (
	"encoding/json"
	"strings"

	"github.com/photoprism/photoprism/pkg/clusters"
)

// Embedding represents a face embedding.
type Embedding []float64

var NullEmbedding = make(Embedding, 512)

// NewEmbedding creates a new embedding from an inference result.
func NewEmbedding(inference []float32) Embedding {
	result := make(Embedding, len(inference))

	var v float32
	var i int

	for i, v = range inference {
		result[i] = float64(v)
	}

	return result
}

// Blacklisted tests if the embedding is blacklisted.
func (m Embedding) Blacklisted() bool {
	return Blacklist.Contains(m, BlacklistRadius)
}

// Distance calculates the distance to another embedding.
func (m Embedding) Distance(other Embedding) float64 {
	return clusters.EuclideanDistance(m, other)
}

// Magnitude returns the embedding vector length (magnitude).
func (m Embedding) Magnitude() float64 {
	return m.Distance(NullEmbedding)
}

// NotBlacklisted tests if the embedding is not blacklisted.
func (m Embedding) NotBlacklisted() bool {
	return !m.Blacklisted()
}

// JSON returns the embedding as JSON bytes.
func (m Embedding) JSON() []byte {
	var noResult = []byte("")

	if len(m) < 1 {
		return noResult
	}

	if result, err := json.Marshal(m); err != nil {
		return noResult
	} else {
		return result
	}
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
