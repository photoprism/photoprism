package face

import (
	"math/rand/v2"
)

// Kind identifies the type of embedding.
type Kind int

// Kind values are stored in faces.face_kind, so 2 and 3 stay reserved for the retired
// children and background classifications rather than being reused.
const (
	// UnclassifiedFace is the kind a cluster is created with. Nothing classifies a face since the
	// child and background filters were removed, so this is what every current cluster holds; it
	// takes part in matching exactly as RegularFace does.
	UnclassifiedFace Kind = 0
	// RegularFace represents a standard face embedding. Written by releases that still classified
	// faces at creation, and by nothing today.
	RegularFace Kind = 1
	// AmbiguousFace represents embeddings that should be treated as uncertain.
	AmbiguousFace Kind = 4
)

// RandomDist returns a distance threshold for matching RandomDEmbeddings.
func RandomDist() float64 {
	return RandomFloat64(0.75, 0.15)
}

// RandomFloat64 adds a random distance offset to a float64.
func RandomFloat64(f, d float64) float64 {
	return f + (rand.Float64()-0.5)*d //nolint:gosec // pseudo-random is sufficient for test fixtures
}

// RandomEmbeddingDims returns the vector length that random embeddings must have to be
// accepted by the configured model, defaulting to the FaceNet length when none is set.
// Persistence rejects vectors of any other length, so fixtures have to follow the model.
func RandomEmbeddingDims() int {
	if dims := ExpectedDims(); dims > 0 {
		return dims
	}

	return 512
}

// RandomEmbeddings returns random embeddings for testing.
func RandomEmbeddings(n int, k Kind) (result Embeddings) {
	if n <= 0 {
		return Embeddings{}
	}

	result = make(Embeddings, n)

	for i := range result {
		result[i] = RandomEmbedding()
	}

	return result
}

// RandomEmbedding returns a random embedding for testing.
func RandomEmbedding() (result Embedding) {
	dims := RandomEmbeddingDims()
	result = make(Embedding, dims)

	d := 64 / float64(dims)

	for i := range result {
		result[i] = RandomFloat64(0, d)
	}

	normalizeEmbedding(result)

	return result
}
