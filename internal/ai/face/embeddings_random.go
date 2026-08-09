package face

import (
	"math/rand/v2"
)

// Kind identifies the type of embedding.
type Kind int

const (
	// RegularFace represents a standard face embedding.
	RegularFace Kind = iota + 1
	// ChildrenFace represents a child face embedding.
	ChildrenFace
	// BackgroundFace represents non-face/background embeddings.
	BackgroundFace
	// AmbiguousFace represents embeddings that should be treated as uncertain.
	AmbiguousFace
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
		switch k {
		case RegularFace:
			result[i] = RandomEmbedding()
		case ChildrenFace:
			result[i] = RandomChildrenEmbedding()
		case BackgroundFace:
			result[i] = RandomBackgroundEmbedding()
		}

	}

	return result
}

// RandomEmbedding returns a random embedding for testing.
func RandomEmbedding() (result Embedding) {
	dims := RandomEmbeddingDims()
	result = make(Embedding, dims)

	d := 64 / float64(dims)

	for {
		i := 0
		for i = range result {
			result[i] = RandomFloat64(0, d)
		}
		if !result.SkipMatching() {
			break
		}
	}

	normalizeEmbedding(result)

	return result
}

// RandomChildrenEmbedding returns a random children embedding for testing.
func RandomChildrenEmbedding() (result Embedding) {
	// The bundled samples are FaceNet-space vectors, so under any other model there is no
	// child region to perturb and a plain random vector of the right length is all we can
	// return. IsChild is inactive there for the same reason.
	if len(Children) == 0 || !SamplesComparable() {
		return RandomEmbedding()
	}

	result = make(Embedding, len(Children[0].Embedding))

	d := 0.1 / float64(len(result))
	n := rand.IntN(len(Children)) //nolint:gosec // deterministic seeding not required for synthetic embeddings
	e := Children[n].Embedding

	for i := range result {
		result[i] = RandomFloat64(e[i], d)
	}

	normalizeEmbedding(result)

	return result
}

// RandomBackgroundEmbedding returns a random background embedding for testing.
func RandomBackgroundEmbedding() (result Embedding) {
	if len(Background) == 0 || !SamplesComparable() {
		return RandomEmbedding()
	}

	result = make(Embedding, len(Background[0].Embedding))

	d := 0.1 / float64(len(result))
	n := rand.IntN(len(Background)) //nolint:gosec // deterministic seeding not required for synthetic embeddings
	e := Background[n].Embedding

	for i := range result {
		result[i] = RandomFloat64(e[i], d)
	}

	normalizeEmbedding(result)

	return result
}
