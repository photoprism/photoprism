package face

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/pkg/clusters"
	"github.com/tidwall/gjson"
	"gonum.org/v1/gonum/mat"
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

// NewEmbeddingFromJsonArray creates a new Embedding from a JSON array
func NewEmbeddingFromJsonArray(array []gjson.Result) Embedding {
	result := make(Embedding, len(array))
	for i, jsonResult := range array {
		result[i] = float64(jsonResult.Num)
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

// CanMatch tests if the face embedding is not blacklisted.
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

// CosineSimilarity calculates the cosine similarity to another embedding.
func (m Embedding) CosineSimilarity(other Embedding) float64 {
	m_vec := mat.NewVecDense(len(m), m)
	other_vec := mat.NewVecDense(len(other), other)
	return mat.Dot(m_vec, other_vec) / (mat.Norm(m_vec, 2) * mat.Norm(other_vec, 2))
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

// Marshal returns an embedding as comma seperated string in braces. Example output: [1.0,2.0,3.0]
func (m Embedding) MarshalEmbedding() string {
	text := "["
	sep := ""
	for _, number := range m {
		text += sep + fmt.Sprintf("%f", number)
		sep = ","
	}
	text += "]"
	return text
}
