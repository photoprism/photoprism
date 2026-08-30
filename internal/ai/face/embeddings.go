package face

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
)

// Embeddings represents a face embedding cluster.
type Embeddings []Embedding

// NewEmbeddings creates a new embeddings from inference results.
func NewEmbeddings(inference [][]float32) Embeddings {
	result := make(Embeddings, len(inference))

	for i, v := range inference {
		result[i] = NewEmbedding(v)
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

// One tests if there is exactly one embedding.
func (embeddings Embeddings) One() bool {
	return embeddings.Count() == 1
}

// ValidEmbeddings checks the cardinality, dimensions, and values of an embedding result.
// Non-finite values survive JSON and storage but poison every later distance, so they are
// rejected where the vector enters the index rather than where it is compared.
func ValidEmbeddings(embeddings Embeddings, dims int) bool {
	if !embeddings.One() || dims < 1 || len(embeddings[0]) != dims {
		return false
	}

	for _, value := range embeddings[0] {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}

	// A vector with no magnitude is not a face, and normalization cannot turn it into one.
	return !embeddings[0].Zero()
}

// Dims returns the number of values shared by all embeddings, 0 when there are none,
// and -1 when they differ, since vectors of different lengths cannot be compared.
func (embeddings Embeddings) Dims() int {
	if len(embeddings) < 1 {
		return 0
	}

	dims := len(embeddings[0])

	for i := 1; i < len(embeddings); i++ {
		if len(embeddings[i]) != dims {
			return -1
		}
	}

	return dims
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
// Dist returns -1 for vectors of a different length, which would otherwise read as the
// closest possible match, so only a non-negative distance can be inside the radius.
func (embeddings Embeddings) Contains(other Embedding, radius float64) bool {
	for _, e := range embeddings {
		if d := e.Dist(other); d >= 0 && d < radius {
			return true
		}
	}

	return false
}

// Dist returns the minimum distance to an embedding, or -1 when none is comparable.
// A vector Embedding.Dist cannot compare is skipped rather than counted: it reports -1,
// which would otherwise win the minimum over every real distance.
func (embeddings Embeddings) Dist(other Embedding) (dist float64) {
	dist = -1

	for _, e := range embeddings {
		if d := e.Dist(other); d >= 0 && (dist < 0 || d < dist) {
			dist = d
		}
	}

	return dist
}

// DistWithin returns the minimum distance to an embedding, or -1 when none is comparable or
// none is within limit. Each hit tightens the limit for the embeddings that follow it, so the
// result is the same minimum Dist would report whenever that minimum is within the limit.
func (embeddings Embeddings) DistWithin(other Embedding, limit float64) (dist float64) {
	dist = -1

	for _, e := range embeddings {
		if d := e.DistWithin(other, limit); d >= 0 && (dist < 0 || d < dist) {
			dist = d
			limit = d
		}
	}

	return dist
}

// JSON returns the embeddings as JSON-encoded bytes.
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

	var first Embedding
	for _, emb := range embeddings {
		first = emb
		break
	}

	// Only one embedding?
	if count == 1 {
		// Return embedding if there is only one.
		return first, 0.0, 1
	}

	dim := len(first)

	// No embedding values?
	if dim == 0 {
		return Embedding{}, 0.0, count
	}

	result = make(Embedding, dim)

	// Vectors of a different length belong to another embedding space, so the mean is
	// scaled by the vectors that actually contributed rather than by all of them.
	contributors := 0

	for i := range embeddings {
		emb := embeddings[i]

		if len(emb) != dim {
			continue
		}

		contributors++

		normalizeEmbedding(emb)

		for j := range dim {
			result[j] += emb[j]
		}
	}

	invCount := 1.0 / float64(contributors)

	for i := range dim {
		result[i] *= invCount
	}

	normalizeEmbedding(result)

	dists := make([]float64, 0, count)

	for _, emb := range embeddings {
		if len(emb) != dim {
			continue
		}

		var dist float64

		for i := range dim {
			diff := result[i] - emb[i]
			dist += diff * diff
		}

		dists = append(dists, math.Sqrt(dist))
	}

	// Epsilon is the tolerance the comparison path uses, so a sample exactly on the radius is still
	// inside it. Raised only when positive: zero means unmeasurable, which SetEmbeddings answers
	// with the full cluster radius, and a floor here would consume that meaning.
	if d := percentileOf(dists, ClusterPercentile); d > 0 {
		radius = d + Epsilon
	}

	return result, radius, count
}

// percentileOf returns the distance at the given percentile by nearest rank, so the result is always
// one of the values passed. It sorts in place, and reports 0 for an empty slice or a percentile at
// or below zero - which would otherwise select the smallest distance and give a cluster no reach.
func percentileOf(dists []float64, p int) float64 {
	if len(dists) == 0 || p < 1 {
		return 0
	}

	slices.Sort(dists)

	rank := min(max((p*len(dists)+99)/100, 1), len(dists))

	return dists[rank-1]
}

// Radius returns how far from their midpoint the ClusterPercentile of the embeddings reach, before
// ClampSampleRadius bounds what a cluster built from them would store. It normalizes its receiver
// in place, as EmbeddingsMidpoint does, so it is not the pure accessor it reads as.
func (embeddings Embeddings) Radius() (radius float64) {
	_, radius, _ = EmbeddingsMidpoint(embeddings)
	return radius
}

// RadiusFrom returns how far from center the ClusterPercentile of the embeddings reach, in the shape
// SetEmbeddings stores, and reports whether every one of them could be measured.
//
// Center is taken as given rather than recomputed, because a cluster's id is the hash of its own
// centroid - deriving a new one here would change its identity and orphan every marker holding it.
func RadiusFrom(center Embedding, embeddings Embeddings) (radius float64, ok bool) {
	if len(center) == 0 || len(embeddings) == 0 {
		return 0, false
	}

	dists := make([]float64, 0, len(embeddings))

	for _, emb := range embeddings {
		// A vector of another width or holding a non-finite component yields -1, and one with no
		// magnitude sits a unit from every unit vector. Neither is a distance, and answering with
		// the widest radius in the schema is what this replaces rather than something to fall back
		// on - so a set holding either is declined whole.
		if d := center.Dist(emb); d < 0 || emb.Zero() {
			return 0, false
		} else {
			dists = append(dists, d)
		}
	}

	if d := percentileOf(dists, ClusterPercentile); d > 0 {
		radius = ClampSampleRadius(d + Epsilon)
	}

	// A spread that measured zero - one member, or copies of one crop - is unmeasurable rather than
	// tight, which SetEmbeddings answers the same way: the full radius, not a cluster narrower than
	// any real pair of one person's faces.
	if radius <= 0 {
		return ClusterRadius, true
	}

	return radius, true
}

// ClusterFits reports whether a cluster of the given radius would accept its own members.
//
// Not implied by ClusterDist: DBSCAN bounds the distance to a neighbor rather than the width of the
// result. Past ClusterRadius the clamped radius stops the gate widening with the group, so anything
// beyond that sum is a member its own cluster would refuse.
func ClusterFits(radius float64) bool {
	return radius <= AcceptDist(radius)
}

// UnmarshalEmbeddings parses face embedding JSON.
func UnmarshalEmbeddings(s string) (result Embeddings, err error) {
	if s == "" {
		return result, fmt.Errorf("cannot unmarshal empeddings, empty string provided")
	} else if !strings.HasPrefix(s, "[[") {
		return result, fmt.Errorf("cannot unmarshal empeddings, invalid json provided")
	}

	err = json.Unmarshal([]byte(s), &result)

	for i := range result {
		normalizeEmbedding(result[i])
	}

	return result, err
}
