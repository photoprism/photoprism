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

// CosineSimilarity calculates the cosine similarity to another embedding.
func (m Embedding) CosineSimilarity(other Embedding) float64 {
	m_vec := mat.NewVecDense(len(m), m)
	other_vec := mat.NewVecDense(len(other), other)
	return mat.Dot(m_vec, other_vec) / (mat.Norm(m_vec, 2) * mat.Norm(other_vec, 2))
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
