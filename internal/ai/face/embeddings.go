package face

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/montanaflynn/stats"

	"github.com/photoprism/photoprism/pkg/clusters"
)

// Embeddings represents a face embedding cluster.
type Embeddings []Embedding

// NewEmbeddings creates a new embeddings from inference results.
func NewEmbeddings(inference [][]float32) Embeddings {
	result := make(Embeddings, len(inference))

	var v []float32
	var i int

	for i, v = range inference {
		e := NewEmbedding(v)

		if e.CanMatch() {
			result[i] = e
		}
	}

	return result
}

// Empty tests if embeddings are empty.
func (embeddings Embeddings) Empty() bool {
	if len(embeddings) < 1 {
		return true
	}

	return len(embeddings[0]) < 1
}

// Count returns the number of embeddings.
func (embeddings Embeddings) Count() int {
	if embeddings.Empty() {
		return 0
	}

	return len(embeddings)
}

// Kind returns the type of face e.g. regular, kids, or ignored.
func (embeddings Embeddings) Kind() (result Kind) {
	for _, e := range embeddings {
		if k := e.Kind(); k > result {
			result = k
		}
	}

	return result
}

// One tests if there is exactly one embedding.
func (embeddings Embeddings) One() bool {
	return embeddings.Count() == 1
}

// First returns the first face embedding.
func (embeddings Embeddings) First() Embedding {
	if embeddings.Empty() {
		return NullEmbedding
	}

	return embeddings[0]
}

// Float64 returns embeddings as a float64 slice.
func (embeddings Embeddings) Float64() [][]float64 {
	result := make([][]float64, len(embeddings))

	for i, e := range embeddings {
		result[i] = e
	}

	return result
}

// Contains tests if another embeddings is contained within a radius.
func (embeddings Embeddings) Contains(other Embedding, radius float64) bool {
	for _, e := range embeddings {
		if d := e.Dist(other); d < radius {
			return true
		}
	}

	return false
}

// Dist returns the minimum distance to an embedding.
func (embeddings Embeddings) Dist(other Embedding) (dist float64) {
	dist = -1

	for _, e := range embeddings {
		if d := e.Dist(other); d < dist || dist < 0 {
			dist = d
		}
	}

	return dist
}

// JSON returns the embeddings as JSON bytes.
func (embeddings Embeddings) JSON() []byte {
	var noResult = []byte("")

	if embeddings.Empty() {
		return noResult
	}

	if result, err := json.Marshal(embeddings); err != nil {
		return noResult
	} else {
		return result
	}
}

// EmbeddingsMidpoint returns the embeddings vector midpoint.
func EmbeddingsMidpoint(embeddings Embeddings) (result Embedding, radius float64, count int) {
	// Return if there are no embeddings.
	if embeddings.Empty() {
		return Embedding{}, 0, 0
	}

	// Count embeddings.
	count = len(embeddings)

	// Only one embedding?
	if count == 1 {
		// Return embedding if there is only one.
		return embeddings[0], 0.0, 1
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
		if d := clusters.EuclideanDist(result, emb); d > radius {
			radius = d + 0.01
		}
	}

	return result, radius, count
}

// UnmarshalEmbeddings parses face embedding JSON.
func UnmarshalEmbeddings(s string) (result Embeddings, err error) {
	if s == "" {
		return result, fmt.Errorf("cannot unmarshal empeddings, empty string provided")
	} else if !strings.HasPrefix(s, "[[") {
		return result, fmt.Errorf("cannot unmarshal empeddings, invalid json provided")
	}

	err = json.Unmarshal([]byte(s), &result)

	return result, err
}
