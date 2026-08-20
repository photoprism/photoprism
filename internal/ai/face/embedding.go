package face

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// Embedding represents a face embedding.
type Embedding []float64

// NullEmbedding is a zero-value placeholder embedding used when no data is available.
var NullEmbedding = make(Embedding, 512)

// NewEmbedding creates a new embedding from an inference result.
func NewEmbedding(inference []float32) Embedding {
	result := make(Embedding, len(inference))

	var v float32
	var i int

	for i, v = range inference {
		result[i] = float64(v)
	}

	normalizeEmbedding(result)

	return result
}

// Dist calculates the distance to another face embedding.
func (m Embedding) Dist(other Embedding) float64 {
	if len(other) == 0 || len(m) != len(other) {
		return -1
	}

	var sum float64

	for i, value := range m {
		diff := value - other[i]
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

// Magnitude returns the face embedding vector length (magnitude).
func (m Embedding) Magnitude() float64 {
	var sum float64

	for _, v := range m {
		sum += v * v
	}

	return math.Sqrt(sum)
}

// normalizeTolerance is how far the sum of squares may sit from 1 for a vector to count
// as already normalized. A unit vector stored as float32 lands within about 1e-7 of it.
const normalizeTolerance = 1e-6

// normalizeEmbedding scales a vector to unit length in place.
//
// Vectors reach this from two directions - freshly unmarshaled ones that are already
// normalized, and raw model output that is not - so an already-unit vector returns
// before the write loop, which is also what keeps it from rewriting the caller's data.
func normalizeEmbedding(e Embedding) {
	var sum float64

	for _, v := range e {
		sum += v * v
	}

	if sum == 0 || math.Abs(sum-1) < normalizeTolerance {
		return
	}

	inv := 1 / math.Sqrt(sum)

	for i := range e {
		e[i] *= inv
	}
}

// JSON returns the face embedding as JSON-encoded bytes.
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

	normalizeEmbedding(result)

	return result, err
}
