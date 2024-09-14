package face

import (
	"encoding/json"
	"fmt"
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

// Kind returns the type of face e.g. regular, kids, or ignored.
func (m Embedding) Kind() Kind {
	if m.KidsFace() {
		return KidsFace
	} else if m.Ignored() {
		return IgnoredFace
	}

	return RegularFace
}

// SkipMatching checks if the face embedding seems unsuitable for matching.
func (m Embedding) SkipMatching() bool {
	return m.KidsFace() || m.Ignored()
}

// CanMatch tests if the face embedding is not excluded.
func (m Embedding) CanMatch() bool {
	return !m.Ignored()
}

// Dist calculates the distance to another face embedding.
func (m Embedding) Dist(other Embedding) float64 {
	if len(other) == 0 || len(m) != len(other) {
		return -1
	}

	return clusters.EuclideanDist(m, other)
}

// Magnitude returns the face embedding vector length (magnitude).
func (m Embedding) Magnitude() float64 {
	return m.Dist(NullEmbedding)
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
func UnmarshalEmbedding(s string) (result Embedding, err error) {
	if s == "" {
		return result, fmt.Errorf("cannot unmarshal embedding, empty string provided")
	} else if !strings.HasPrefix(s, "[") {
		return result, fmt.Errorf("cannot unmarshal embedding, invalid json provided")
	}

	err = json.Unmarshal([]byte(s), &result)

	return result, err
}
