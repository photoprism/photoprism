package face

import (
	"math"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// calibrationSample holds one identity's cluster centroid together with the queries
// that must be matched against it. It mirrors an indexed face cluster: a midpoint
// built from a few samples, and further markers matched against that midpoint.
type calibrationSample struct {
	subject  string
	centroid Embedding
	radius   float64
	queries  []Embedding
}

// buildCalibrationSamples splits each identity into cluster samples and queries.
// The centroid comes from the function the indexer uses, so the measured distances
// are the ones production compares against rather than an approximation of them.
func buildCalibrationSamples(subjects []benchmarkSubject, core int) []calibrationSample {
	if core < 1 {
		core = 1
	}

	var result []calibrationSample

	for _, s := range subjects {
		// An identity needs enough images to seed a cluster and still leave a query.
		if len(s.embeddings) <= core {
			continue
		}

		cluster := make(Embeddings, core)
		copy(cluster, s.embeddings[:core])

		centroid, radius, _ := EmbeddingsMidpoint(cluster)

		if len(centroid) == 0 {
			continue
		}

		queries := make([]Embedding, len(s.embeddings)-core)
		copy(queries, s.embeddings[core:])

		result = append(result, calibrationSample{subject: s.name, centroid: centroid, radius: radius, queries: queries})
	}

	return result
}

// centroidMargin returns the distance of a query to a cluster centroid minus the
// radius that would be stored for it, which is the quantity Face.Match compares
// against MatchDist. Sweeping a threshold over it therefore yields MatchDist itself.
func centroidMargin(query, centroid Embedding, radius, radiusCap float64) (float64, bool) {
	d := query.Dist(centroid)

	if d < 0 {
		return 0, false
	}

	if radius > radiusCap {
		radius = radiusCap
	}

	return d - radius, true
}

// centroidMarginPairs scores every query against its own centroid and a deterministic
// sample of foreign centroids. Negatives are strided rather than sampled at random so
// a rerun reproduces the same numbers.
func centroidMarginPairs(samples []calibrationSample, radiusCap float64, maxNegatives int) []scoredPair {
	var pairs []scoredPair

	for i := range samples {
		for _, q := range samples[i].queries {
			if m, ok := centroidMargin(q, samples[i].centroid, samples[i].radius, radiusCap); ok {
				pairs = append(pairs, scoredPair{dist: m, same: true})
			}
		}
	}

	var candidates int

	for i := range samples {
		for j := range samples {
			if i != j {
				candidates += len(samples[j].queries)
			}
		}
	}

	stride := 1

	if maxNegatives > 0 && candidates > maxNegatives {
		stride = candidates / maxNegatives
	}

	counter := 0

	for i := range samples {
		for j := range samples {
			if i == j {
				continue
			}

			for _, q := range samples[j].queries {
				if counter%stride == 0 {
					if m, ok := centroidMargin(q, samples[i].centroid, samples[i].radius, radiusCap); ok {
						pairs = append(pairs, scoredPair{dist: m, same: false})
					}
				}

				counter++
			}
		}
	}

	return pairs
}

// rateAtThreshold returns the true and false accept rates that result from accepting
// every pair at or below the threshold.
func rateAtThreshold(pairs []scoredPair, threshold float64) (tar, far float64) {
	var same, diff, acceptedSame, acceptedDiff float64

	for _, p := range pairs {
		if p.same {
			same++

			if p.dist <= threshold {
				acceptedSame++
			}
		} else {
			diff++

			if p.dist <= threshold {
				acceptedDiff++
			}
		}
	}

	if same > 0 {
		tar = acceptedSame / same
	}

	if diff > 0 {
		far = acceptedDiff / diff
	}

	return tar, far
}

// thresholdAtFAR returns the most permissive threshold whose false accept rate stays
// within the budget, and the true accept rate it reaches there. It reports false when
// the budget cannot be met, which happens when the closest pair is already a false
// accept: translating a threshold into that model's scale is then not meaningful.
func thresholdAtFAR(pairs []scoredPair, budget float64) (threshold, tar float64, ok bool) {
	sorted := make([]scoredPair, len(pairs))
	copy(sorted, pairs)

	sort.Slice(sorted, func(i, j int) bool { return sorted[i].dist < sorted[j].dist })

	var totalSame, totalDiff float64

	for _, p := range sorted {
		if p.same {
			totalSame++
		} else {
			totalDiff++
		}
	}

	if totalSame == 0 || totalDiff == 0 {
		return 0, 0, false
	}

	var acceptedSame, acceptedDiff float64

	// Walk tie groups so a threshold is only taken once every equal score is included,
	// which is what accepting "at or below" that distance would really do.
	for i := 0; i < len(sorted); {
		j := i
		var sameInGroup, diffInGroup float64

		for j < len(sorted) && sorted[j].dist == sorted[i].dist {
			if sorted[j].same {
				sameInGroup++
			} else {
				diffInGroup++
			}

			j++
		}

		if (acceptedDiff+diffInGroup)/totalDiff > budget {
			break
		}

		acceptedSame += sameInGroup
		acceptedDiff += diffInGroup
		threshold = sorted[i].dist
		tar = acceptedSame / totalSame
		ok = true

		i = j
	}

	return threshold, tar, ok
}

// percentile returns the nearest-rank value at the requested percentile.
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	idx := int(math.Round(p * float64(len(sorted)-1)))

	return sorted[max(min(idx, len(sorted)-1), 0)]
}

func TestBuildCalibrationSamples(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		samples := buildCalibrationSamples(testSubjects(3, 5), 4)
		require.Len(t, samples, 3)

		for _, s := range samples {
			assert.Len(t, s.queries, 1)
			assert.NotEmpty(t, s.centroid)
			assert.GreaterOrEqual(t, s.radius, 0.0)
		}
	})
	t.Run("TooFewImages", func(t *testing.T) {
		// Four images seed the core but leave nothing to match against it.
		assert.Empty(t, buildCalibrationSamples(testSubjects(3, 4), 4))
	})
	t.Run("CoreClamped", func(t *testing.T) {
		samples := buildCalibrationSamples(testSubjects(1, 2), 0)
		require.Len(t, samples, 1)
		assert.Len(t, samples[0].queries, 1)
	})
	t.Run("NoSubjects", func(t *testing.T) {
		assert.Empty(t, buildCalibrationSamples(nil, 4))
	})
}

func TestCentroidMargin(t *testing.T) {
	a := Embedding{1, 0}
	b := Embedding{0, 1}

	t.Run("Success", func(t *testing.T) {
		m, ok := centroidMargin(a, b, 0.2, 0.42)
		require.True(t, ok)
		assert.InDelta(t, math.Sqrt2-0.2, m, 0.0001)
	})
	t.Run("RadiusCapped", func(t *testing.T) {
		m, ok := centroidMargin(a, b, 0.9, 0.42)
		require.True(t, ok)
		assert.InDelta(t, math.Sqrt2-0.42, m, 0.0001)
	})
	t.Run("DimensionMismatch", func(t *testing.T) {
		_, ok := centroidMargin(a, Embedding{0, 1, 0}, 0.2, 0.42)
		assert.False(t, ok)
	})
}

func TestCentroidMarginPairs(t *testing.T) {
	t.Run("AllPairs", func(t *testing.T) {
		samples := buildCalibrationSamples(testSubjects(3, 5), 4)
		pairs := centroidMarginPairs(samples, ClusterRadius, 0)

		var same, diff int

		for _, p := range pairs {
			if p.same {
				same++
			} else {
				diff++
			}
		}

		// Three identities with one query each: 3 positives and 3*2 negatives.
		assert.Equal(t, 3, same)
		assert.Equal(t, 6, diff)
	})
	t.Run("NegativesCapped", func(t *testing.T) {
		samples := buildCalibrationSamples(testSubjects(6, 6), 4)
		pairs := centroidMarginPairs(samples, ClusterRadius, 4)

		var diff int

		for _, p := range pairs {
			if !p.same {
				diff++
			}
		}

		assert.LessOrEqual(t, diff, 8)
		assert.Positive(t, diff)
	})
	t.Run("Deterministic", func(t *testing.T) {
		samples := buildCalibrationSamples(testSubjects(5, 6), 4)
		assert.Equal(t, centroidMarginPairs(samples, ClusterRadius, 4), centroidMarginPairs(samples, ClusterRadius, 4))
	})
	t.Run("NoSamples", func(t *testing.T) {
		assert.Empty(t, centroidMarginPairs(nil, ClusterRadius, 0))
	})
}

func TestRateAtThreshold(t *testing.T) {
	pairs := []scoredPair{
		{dist: 0.1, same: true},
		{dist: 0.7, same: true},
		{dist: 0.4, same: false},
		{dist: 1.2, same: false},
	}
	t.Run("Midpoint", func(t *testing.T) {
		tar, far := rateAtThreshold(pairs, 0.5)
		assert.InDelta(t, 0.5, tar, 0.0001)
		assert.InDelta(t, 0.5, far, 0.0001)
	})
	t.Run("AcceptsAll", func(t *testing.T) {
		tar, far := rateAtThreshold(pairs, 2)
		assert.InDelta(t, 1, tar, 0.0001)
		assert.InDelta(t, 1, far, 0.0001)
	})
	t.Run("AcceptsNone", func(t *testing.T) {
		tar, far := rateAtThreshold(pairs, 0)
		assert.InDelta(t, 0, tar, 0.0001)
		assert.InDelta(t, 0, far, 0.0001)
	})
	t.Run("Empty", func(t *testing.T) {
		tar, far := rateAtThreshold(nil, 0.5)
		assert.InDelta(t, 0, tar, 0.0001)
		assert.InDelta(t, 0, far, 0.0001)
	})
}

func TestThresholdAtFAR(t *testing.T) {
	t.Run("CleanSeparation", func(t *testing.T) {
		pairs := []scoredPair{
			{dist: 0.1, same: true},
			{dist: 0.2, same: true},
			{dist: 0.9, same: false},
			{dist: 1.0, same: false},
		}
		d, tar, ok := thresholdAtFAR(pairs, 0)
		require.True(t, ok)
		assert.InDelta(t, 0.2, d, 0.0001)
		assert.InDelta(t, 1, tar, 0.0001)
	})
	t.Run("BudgetSpent", func(t *testing.T) {
		pairs := []scoredPair{
			{dist: 0.1, same: true},
			{dist: 0.3, same: false},
			{dist: 0.5, same: true},
			{dist: 0.9, same: false},
		}
		// A budget of half the negatives reaches the second positive.
		d, tar, ok := thresholdAtFAR(pairs, 0.5)
		require.True(t, ok)
		assert.InDelta(t, 0.5, d, 0.0001)
		assert.InDelta(t, 1, tar, 0.0001)
	})
	t.Run("BudgetUnreachable", func(t *testing.T) {
		pairs := []scoredPair{
			{dist: 0.1, same: false},
			{dist: 0.5, same: true},
		}
		_, _, ok := thresholdAtFAR(pairs, 0)
		assert.False(t, ok)
	})
	t.Run("Ties", func(t *testing.T) {
		pairs := []scoredPair{
			{dist: 0.5, same: true},
			{dist: 0.5, same: false},
		}
		// The tie group cannot be split, so a zero budget rejects both.
		_, _, ok := thresholdAtFAR(pairs, 0)
		assert.False(t, ok)
	})
	t.Run("NoNegatives", func(t *testing.T) {
		_, _, ok := thresholdAtFAR([]scoredPair{{dist: 0.2, same: true}}, 0.5)
		assert.False(t, ok)
	})
	t.Run("Empty", func(t *testing.T) {
		_, _, ok := thresholdAtFAR(nil, 0.5)
		assert.False(t, ok)
	})
}

func TestCandidateRadiusCaps(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		caps := candidateRadiusCaps([]float64{0.10, 0.20, 0.30, 0.40, 0.50})
		assert.Contains(t, caps, ClusterRadius)
		assert.True(t, sort.Float64sAreSorted(caps))

		for i := 1; i < len(caps); i++ {
			assert.NotEqual(t, caps[i-1], caps[i])
		}
	})
	t.Run("NoRadii", func(t *testing.T) {
		assert.Equal(t, []float64{ClusterRadius}, candidateRadiusCaps(nil))
	})
	t.Run("ZeroRadii", func(t *testing.T) {
		// Single-image clusters have no radius and must not add a zero candidate.
		assert.Equal(t, []float64{ClusterRadius}, candidateRadiusCaps([]float64{0, 0, 0}))
	})
}

// calibrationTestMargins returns margin pairs for two candidate radius caps, where the
// wider cap shifts every score down by the same amount.
func calibrationTestMargins() (map[float64][]scoredPair, []float64) {
	return map[float64][]scoredPair{
		ClusterRadius: {
			{dist: 0.1, same: true},
			{dist: 0.8, same: true},
			{dist: 1.6, same: false},
			{dist: 1.9, same: false},
		},
		0.6: {
			{dist: 0.0, same: true},
			{dist: 0.6, same: true},
			{dist: 1.4, same: false},
			{dist: 1.7, same: false},
		},
	}, []float64{ClusterRadius, 0.6}
}

func TestBestOperatingPoint(t *testing.T) {
	margin, caps := calibrationTestMargins()

	t.Run("Success", func(t *testing.T) {
		p := bestOperatingPoint(margin, caps, 0)
		require.True(t, p.ok)
		assert.InDelta(t, 1, p.tar, 0.0001)
		assert.InDelta(t, 0, p.far, 0.0001)
		// Ties keep the smaller cap, so the shipped value wins when both reach TAR 1.
		assert.InDelta(t, ClusterRadius, p.clusterRadius, 0.0001)
		assert.InDelta(t, 0.8, p.matchDist, 0.0001)
	})
	t.Run("SpendsBudget", func(t *testing.T) {
		p := bestOperatingPoint(margin, caps, 0.5)
		require.True(t, p.ok)
		assert.InDelta(t, 1.6, p.matchDist, 0.0001)
		assert.InDelta(t, 0.5, p.far, 0.0001)
	})
	t.Run("BudgetUnreachable", func(t *testing.T) {
		impossible := map[float64][]scoredPair{ClusterRadius: {{dist: 0.1, same: false}, {dist: 0.5, same: true}}}
		p := bestOperatingPoint(impossible, []float64{ClusterRadius}, 0)
		assert.False(t, p.ok)
		// The shipped values are retained so the report never suggests a blank constant.
		assert.InDelta(t, ClusterRadius, p.clusterRadius, 0.0001)
		assert.InDelta(t, MatchDist, p.matchDist, 0.0001)
	})
}

func TestCalibrateModel(t *testing.T) {
	// A model whose distances are twice the baseline's needs twice the threshold.
	pairwise := []scoredPair{
		{dist: 0.2, same: true},
		{dist: 1.0, same: true},
		{dist: 1.8, same: false},
		{dist: 2.0, same: false},
	}
	margin, caps := calibrationTestMargins()

	t.Run("Success", func(t *testing.T) {
		r := calibrateModel(ModelSFace, pairwise, margin, caps, 0, 0.5)
		assert.Equal(t, ModelSFace, r.model)
		require.True(t, r.clusterDistOK)
		assert.InDelta(t, 1.0, r.clusterDist, 0.0001)
		require.True(t, r.matched.ok)
		assert.InDelta(t, 0.5, r.matched.far, 0.0001)
	})
	t.Run("StricterPointSpendsLess", func(t *testing.T) {
		r := calibrateModel(ModelSFace, pairwise, margin, caps, 0, 0.5)
		require.True(t, r.strict.ok)
		assert.Less(t, r.strict.far, r.matched.far)
		assert.LessOrEqual(t, r.strict.tar, r.matched.tar)
	})
	t.Run("ReportsCurrentBehavior", func(t *testing.T) {
		r := calibrateModel(ModelSFace, pairwise, margin, caps, 0, 0)
		// Today's MatchDist of 0.4 only accepts the closer of the two true pairs.
		assert.InDelta(t, 0.5, r.currentTAR, 0.0001)
		assert.InDelta(t, 0, r.currentFAR, 0.0001)
	})
	t.Run("PairwiseBudgetUnreachable", func(t *testing.T) {
		r := calibrateModel(ModelSFace, []scoredPair{{dist: 0.1, same: false}, {dist: 0.5, same: true}}, margin, caps, 0, 0)
		assert.False(t, r.clusterDistOK)
	})
}

func TestReportCalibration(t *testing.T) {
	result := calibrationResult{
		model:         ModelSFace,
		clusterDist:   1.05,
		clusterDistOK: true,
		currentTAR:    0.58,
		currentFAR:    0.0,
		matched:       operatingPoint{clusterRadius: 0.6, matchDist: 0.72, tar: 0.95, far: 0.0143, ok: true},
		strict:        operatingPoint{clusterRadius: 0.42, matchDist: 0.61, tar: 0.91, far: 0.0014, ok: true},
	}
	t.Run("Success", func(t *testing.T) {
		out := reportCalibration("/tmp/faces", []string{"general"}, 0.0001, 0.0143, []calibrationResult{result})
		assert.Contains(t, out, "Face Threshold Calibration")
		assert.Contains(t, out, ModelSFace)
		assert.Contains(t, out, "1.050")
		assert.Contains(t, out, "0.720")
		assert.Contains(t, out, "0.0014")
		assert.Contains(t, out, "general")
		assert.Contains(t, out, "Budget Matched")
	})
	t.Run("Unavailable", func(t *testing.T) {
		out := reportCalibration("/tmp/faces", nil, 0, 0, []calibrationResult{{model: ModelSFace}})
		assert.Contains(t, out, "n/a")
	})
}

func TestWriteOperatingPoints(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var b strings.Builder
		writeOperatingPoints(&b, []calibrationResult{{
			model:  ModelSFace,
			strict: operatingPoint{clusterRadius: 0.42, matchDist: 0.61, tar: 0.91, far: 0.0014, ok: true},
		}}, func(r calibrationResult) operatingPoint { return r.strict })
		assert.Contains(t, b.String(), "0.610")
		assert.Contains(t, b.String(), "0.9100")
	})
	t.Run("Unavailable", func(t *testing.T) {
		var b strings.Builder
		writeOperatingPoints(&b, []calibrationResult{{model: ModelSFace}}, func(r calibrationResult) operatingPoint { return r.matched })
		assert.Contains(t, b.String(), "n/a")
	})
}

func TestPercentile(t *testing.T) {
	values := []float64{0.4, 0.1, 0.3, 0.2, 0.5}
	t.Run("Median", func(t *testing.T) {
		assert.InDelta(t, 0.3, percentile(values, 0.5), 0.0001)
	})
	t.Run("Min", func(t *testing.T) {
		assert.InDelta(t, 0.1, percentile(values, 0), 0.0001)
	})
	t.Run("Max", func(t *testing.T) {
		assert.InDelta(t, 0.5, percentile(values, 1), 0.0001)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.InDelta(t, 0, percentile(nil, 0.5), 0.0001)
	})
}
