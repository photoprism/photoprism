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

// Blacklisted tests if the face embedding is blacklisted.
func (m Embedding) Blacklisted() bool {
	return Blacklist.Contains(m, BlacklistRadius)
}

// Child tests if the face embedding belongs to a child.
func (m Embedding) Child() bool {
	return Children.Contains(m, ChildrenRadius)
}

// Unsuitable tests if the face embedding is unsuitable for clustering and matching.
func (m Embedding) Unsuitable() bool {
	return m.Child() || m.Blacklisted()
}

// Distance calculates the distance to another face embedding.
func (m Embedding) Distance(other Embedding) float64 {
	return clusters.EuclideanDistance(m, other)
}

// Magnitude returns the face embedding vector length (magnitude).
func (m Embedding) Magnitude() float64 {
	return m.Distance(NullEmbedding)
}

// NotBlacklisted tests if the face embedding is not blacklisted.
func (m Embedding) NotBlacklisted() bool {
	return !m.Blacklisted()
}

// JSON returns the face embedding as JSON bytes.
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
