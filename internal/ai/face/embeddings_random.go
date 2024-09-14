package face

import (
	"math/rand/v2"
)

type Kind int

const (
	RegularFace Kind = iota + 1
	KidsFace
	IgnoredFace
	AmbiguousFace
)

// RandomDist returns a distance threshold for matching RandomDEmbeddings.
func RandomDist() float64 {
	return RandomFloat64(0.75, 0.15)
}

// RandomFloat64 adds a random distance offset to a float64.
func RandomFloat64(f, d float64) float64 {
	return f + (rand.Float64()-0.5)*d
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
		case KidsFace:
			result[i] = RandomKidsEmbedding()
		case IgnoredFace:
			result[i] = RandomIgnoredEmbedding()
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

	return result
}

// RandomKidsEmbedding returns a random kids embedding for testing.
func RandomKidsEmbedding() (result Embedding) {
	result = make(Embedding, 512)

	d := 0.1 / 512.0
	n := 1 + rand.IntN(len(KidsEmbeddings)-1)
	e := KidsEmbeddings[n]

	for i := range result {
		result[i] = RandomFloat64(e[i], d)
	}

	return result
}

// RandomIgnoredEmbedding returns a random ignored embedding for testing.
func RandomIgnoredEmbedding() (result Embedding) {
	result = make(Embedding, 512)

	d := 0.1 / 512.0
	n := 1 + rand.IntN(len(IgnoredEmbeddings)-1)
	e := IgnoredEmbeddings[n]

	for i := range result {
		result[i] = RandomFloat64(e[i], d)
	}

	return result
}
