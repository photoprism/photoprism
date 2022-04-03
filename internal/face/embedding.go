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

// IgnoreFace tests whether the embedding is generally unsuitable for matching.
func (m Embedding) IgnoreFace() bool {
	if IgnoreDist <= 0 {
		return false
	}

	return IgnoreEmbeddings.Contains(m, IgnoreDist)
}

// KidsFace tests if the embedded face belongs to a baby or young child.
func (m Embedding) KidsFace() bool {
	if KidsDist <= 0 {
		return false
	}

	return KidsEmbeddings.Contains(m, KidsDist)
}

// OmitMatch tests if the face embedding is unsuitable for matching.
func (m Embedding) OmitMatch() bool {
	return m.KidsFace() || m.IgnoreFace()
}

// CanMatch tests if the face embedding is not blacklisted.
func (m Embedding) CanMatch() bool {
	return !m.IgnoreFace()
}

// Dist calculates the distance to another face embedding.
func (m Embedding) Dist(other Embedding) float64 {
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
