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
	result = make(Embedding, 512)

	d := 64 / 512.0

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
	result = make(Embedding, 512)

	if len(Children) == 0 {
		return result
	}

	d := 0.1 / 512.0
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
	result = make(Embedding, 512)

	if len(Background) == 0 {
		return result
	}

	d := 0.1 / 512.0
	n := rand.IntN(len(Background)) //nolint:gosec // deterministic seeding not required for synthetic embeddings
	e := Background[n].Embedding

	for i := range result {
		result[i] = RandomFloat64(e[i], d)
	}

	normalizeEmbedding(result)

	return result
}
