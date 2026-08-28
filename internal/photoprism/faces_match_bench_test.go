package photoprism

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
)

// benchmarkCandidateCount is the number of face clusters a benchmark matches against. Libraries
// of ~125k photos have been reported with ~1.7k clusters, so this is a large but realistic run.
const benchmarkCandidateCount = 2048

// benchmarkEmbeddingAt returns a unit embedding at approximately the given distance from base.
//
// Independent random vectors are near-orthogonal, so a candidate list built from them sits far
// outside every accept distance and costs nothing to reject. Real clusters are spread across the
// whole range, which is what decides how much work a bounded scan can actually skip.
func benchmarkEmbeddingAt(base face.Embedding, dist float64, rnd *rand.Rand) face.Embedding {
	// A chord of length dist on the unit sphere subtends this angle.
	theta := 2 * math.Asin(min(dist/2, 1))

	dir := make(face.Embedding, len(base))

	var dot float64

	for i := range dir {
		dir[i] = rnd.NormFloat64()
		dot += dir[i] * base[i]
	}

	// Remove the base component so the rotation angle is exactly theta.
	var norm float64

	for i := range dir {
		dir[i] -= dot * base[i]
		norm += dir[i] * dir[i]
	}

	if norm == 0 {
		return base
	}

	norm = math.Sqrt(norm)
	result := make(face.Embedding, len(base))

	for i := range result {
		result[i] = base[i]*math.Cos(theta) + (dir[i]/norm)*math.Sin(theta)
	}

	return result
}

// benchmarkFaceIndex builds a candidate list spread over the given distances from a shared base,
// and returns the index together with that base.
func benchmarkFaceIndex(dists []float64) (faceIndex, face.Embedding) {
	rnd := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // deterministic fixtures, not security
	base := face.RandomEmbedding()
	faces := make(entity.Faces, 0, benchmarkCandidateCount)

	for i := range benchmarkCandidateCount {
		emb := benchmarkEmbeddingAt(base, dists[i%len(dists)], rnd)
		f := entity.NewFace("", entity.SrcAuto, face.Embeddings{emb}, face.EmbeddingModelName())
		faces = append(faces, *f)
	}

	return buildFaceIndex(faces), base
}

// BenchmarkSelectBestFace measures selection when the marker has a close cluster to find.
func BenchmarkSelectBestFace(b *testing.B) {
	// A realistic mix: a few clusters near the marker, the rest spread outward.
	index, base := benchmarkFaceIndex([]float64{0.3, 0.6, 0.9, 1.1, 1.2, 1.3, 1.4, 1.4})

	rnd := rand.New(rand.NewPCG(3, 4)) //nolint:gosec // deterministic fixtures, not security
	markerEmb := face.Embeddings{benchmarkEmbeddingAt(base, 0.1, rnd)}

	for b.Loop() {
		selectBestFace(markerEmb, index)
	}

	// b.Loop() resets the timer and the extra metrics on its first iteration, so a metric
	// reported before it never reaches the output.
	b.ReportMetric(float64(len(index.candidates)), "candidates")
}

// BenchmarkSelectBestFaceUnmatched measures selection when nothing matches, which is the case
// that dominates a large library and the one a partitioned search cannot help with, since a
// marker that matches nothing has to be compared with everything either way.
func BenchmarkSelectBestFaceUnmatched(b *testing.B) {
	index, base := benchmarkFaceIndex([]float64{1.3, 1.35, 1.4, 1.41})

	rnd := rand.New(rand.NewPCG(5, 6)) //nolint:gosec // deterministic fixtures, not security
	markerEmb := face.Embeddings{benchmarkEmbeddingAt(base, 1.41, rnd)}

	for b.Loop() {
		selectBestFace(markerEmb, index)
	}

	// b.Loop() resets the timer and the extra metrics on its first iteration, so a metric
	// reported before it never reaches the output.
	b.ReportMetric(float64(len(index.candidates)), "candidates")
}

// BenchmarkSelectBestFaceWinnerFirst and BenchmarkSelectBestFaceWinnerLast measure how much
// candidate order costs. The bound tightens as soon as a close candidate is seen, so meeting
// the winner early makes every later candidate cheaper to reject - which is why the face query
// returns the largest clusters first.
func BenchmarkSelectBestFaceWinnerFirst(b *testing.B) {
	benchmarkSelectBestFaceAt(b, 0)
}

func BenchmarkSelectBestFaceWinnerLast(b *testing.B) {
	benchmarkSelectBestFaceAt(b, benchmarkCandidateCount-1)
}

// benchmarkSelectBestFaceAt places the only close candidate at the given position.
func benchmarkSelectBestFaceAt(b *testing.B, pos int) {
	index, base := benchmarkFaceIndex([]float64{1.2, 1.25, 1.3, 1.35})

	rnd := rand.New(rand.NewPCG(9, 10)) //nolint:gosec // deterministic fixtures, not security
	markerEmb := face.Embeddings{benchmarkEmbeddingAt(base, 0.1, rnd)}

	// One candidate the marker actually matches, everything else far away.
	winner := entity.NewFace("", entity.SrcAuto, markerEmb, face.EmbeddingModelName())
	require.NotNil(b, winner)
	index.candidates[pos] = faceCandidate{ref: winner, emb: winner.Embedding(), acceptDist: winner.AcceptDist()}

	for b.Loop() {
		selectBestFace(markerEmb, index)
	}
}

// BenchmarkSelectBestFaceLegacy captures an unbounded scan over the same candidates, which is
// what selection costs without the per-candidate distance limit.
func BenchmarkSelectBestFaceLegacy(b *testing.B) {
	index, base := benchmarkFaceIndex([]float64{0.3, 0.6, 0.9, 1.1, 1.2, 1.3, 1.4, 1.4})

	rnd := rand.New(rand.NewPCG(3, 4)) //nolint:gosec // deterministic fixtures, not security
	markerEmb := face.Embeddings{benchmarkEmbeddingAt(base, 0.1, rnd)}

	for b.Loop() {
		legacySelectBestFace(markerEmb, index)
	}

	b.ReportMetric(float64(len(index.candidates)), "candidates")
}

// legacySelectBestFace scans every candidate to completion, without abandoning the ones that
// cannot win. It is the reference the bounded scan is measured and verified against, so it
// spells out both gates rather than calling limit(), which is itself under test.
func legacySelectBestFace(embeddings face.Embeddings, idx faceIndex) (*entity.Face, float64) {
	var best *entity.Face

	bestDist := -1.0

	for i := range idx.candidates {
		c := &idx.candidates[i]

		dist := embeddings.Dist(c.emb)

		if dist < 0 || dist > c.acceptDist {
			continue
		}

		if c.collisionRadius > face.CollisionDist && dist > c.collisionRadius {
			continue
		}

		if best == nil || dist < bestDist {
			best = c.ref
			bestDist = dist
		}
	}

	return best, bestDist
}
