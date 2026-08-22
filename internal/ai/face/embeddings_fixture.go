package face

import (
	"math"
	"math/rand/v2"
)

// fixtureSeedStream separates the two counters the fixture generators seed their random
// source with, so vectors generated from the same seed by different functions differ.
const fixtureSeedStream = 0x9e3779b97f4a7c15

// FixtureEmbedding returns a deterministic unit embedding of the length the configured model
// produces, so a fixture regenerated for another model keeps its geometry while adopting that
// model's width. Independent seeds are near-orthogonal, which is what makes two of them stand
// for two different people.
func FixtureEmbedding(seed uint64) Embedding {
	dims := RandomEmbeddingDims()

	if dims < 1 {
		return Embedding{}
	}

	rnd := rand.New(rand.NewPCG(seed, fixtureSeedStream)) //nolint:gosec // deterministic fixtures, not security
	result := make(Embedding, dims)

	for i := range result {
		result[i] = rnd.NormFloat64()
	}

	normalizeEmbedding(result)

	return result
}

// FixtureEmbeddingAt returns a deterministic unit embedding at the given distance from base,
// which is how a fixture states that a marker belongs to a cluster in the distance scale of
// whichever model is configured.
//
// Distances between unit vectors are chords, so the result is base rotated by the angle that
// chord subtends. A distance of zero or less returns a copy of base, and 2 its antipode.
func FixtureEmbeddingAt(base Embedding, dist float64, seed uint64) Embedding {
	result := make(Embedding, len(base))
	copy(result, base)

	if len(base) == 0 || dist <= 0 {
		return result
	}

	theta := 2 * math.Asin(min(dist/2, 1))

	rnd := rand.New(rand.NewPCG(seed, fixtureSeedStream)) //nolint:gosec // deterministic fixtures, not security
	dir := make(Embedding, len(base))

	var dot float64

	for i, v := range base {
		dir[i] = rnd.NormFloat64()
		dot += dir[i] * v
	}

	// Removing the base component leaves a direction orthogonal to it, so rotating by theta
	// moves exactly that far rather than somewhat less.
	var norm float64

	for i, v := range base {
		dir[i] -= dot * v
		norm += dir[i] * dir[i]
	}

	if norm == 0 {
		return result
	}

	norm = math.Sqrt(norm)

	for i, v := range base {
		result[i] = v*math.Cos(theta) + (dir[i]/norm)*math.Sin(theta)
	}

	return result
}
