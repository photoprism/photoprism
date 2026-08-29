package alg

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHDBSCAN(t *testing.T) {
	t.Run("EmptySet", func(t *testing.T) {
		_, err := HDBSCAN(nil, 5, 5, 1, EuclideanDist)
		assert.ErrorIs(t, err, errEmptySet)
	})
	t.Run("RaggedData", func(t *testing.T) {
		_, err := HDBSCAN([][]float64{{1, 2}, {3}}, 2, 2, 1, EuclideanDist)
		assert.ErrorIs(t, err, errRaggedData)
	})
	t.Run("ZeroMinPts", func(t *testing.T) {
		_, err := HDBSCAN([][]float64{{1}, {2}}, 0, 2, 1, EuclideanDist)
		assert.ErrorIs(t, err, errZeroMinpts)
	})
	t.Run("NegativeWorkers", func(t *testing.T) {
		_, err := HDBSCAN([][]float64{{1}, {2}}, 2, 2, -1, EuclideanDist)
		assert.ErrorIs(t, err, errZeroWorkers)
	})
	t.Run("SinglePoint", func(t *testing.T) {
		// One point has no hierarchy, and reporting it as a cluster would say nothing.
		h, err := HDBSCAN([][]float64{{1, 2}}, 5, 5, 1, EuclideanDist)
		assert.NoError(t, err)
		assert.Equal(t, Labels{Noise}, h.Labels())
		assert.Equal(t, []float64{0}, h.Probabilities())
	})
	t.Run("DefaultDistance", func(t *testing.T) {
		h, err := HDBSCAN(blobs(t), 5, 5, 1, nil)
		assert.NoError(t, err)
		assert.Equal(t, 3, h.Labels().Count())
	})
	t.Run("MinSizeNeverBelowMinPts", func(t *testing.T) {
		h, err := HDBSCAN(blobs(t), 5, 1, 1, EuclideanDist)
		assert.NoError(t, err)
		assert.Equal(t, 5, h.MinSize)
	})
	t.Run("SeparatesWellSeparatedGroups", func(t *testing.T) {
		data := blobs(t)
		truth := make([]int, len(data))
		for i := range truth {
			truth[i] = i/30 + 1
		}

		h, err := HDBSCAN(data, 5, 5, 4, EuclideanDist)
		assert.NoError(t, err)
		assert.True(t, separatesGroups(h.Labels(), truth, 3, 28))
	})
	t.Run("WorkersDoNotChangeTheResult", func(t *testing.T) {
		data := blobs(t)
		one, err := HDBSCAN(data, 5, 5, 1, EuclideanDist)
		assert.NoError(t, err)
		many, err := HDBSCAN(data, 5, 5, 8, EuclideanDist)
		assert.NoError(t, err)
		assert.Equal(t, one.Labels(), many.Labels())
		assert.Equal(t, one.Probabilities(), many.Probabilities())
	})
	t.Run("CoincidentPointsStayFinite", func(t *testing.T) {
		// Duplicate faces split at an infinite density level, which the scores have to absorb
		// rather than propagate: a NaN probability would be stored as one.
		data := make([][]float64, 0, 40)
		for range 20 {
			data = append(data, []float64{0, 0}, []float64{5, 5})
		}

		h, err := HDBSCAN(data, 5, 5, 1, EuclideanDist)
		assert.NoError(t, err)
		assert.True(t, finite(h.Probabilities()), "a probability may not be NaN or infinite")
		assert.True(t, finite(h.Outliers()), "an outlier score may not be NaN or infinite")
	})
}

// TestHierarchySeparatesVariableDensity is the property the evaluation rests on: groups of
// different densities are separated in one pass, with no distance named anywhere.
func TestHierarchySeparatesVariableDensity(t *testing.T) {
	data, truth := variableDensity(t)

	h, err := HDBSCAN(data, 5, 5, 4, EuclideanDist)
	assert.NoError(t, err)

	t.Run("SeparatesAllThreeGroups", func(t *testing.T) {
		assert.True(t, separatesGroups(h.Labels(), truth, 3, 32))
	})
	t.Run("NoDistanceCutAchievesIt", func(t *testing.T) {
		// The same hierarchy cut at one level cannot do it, which isolates the extraction as
		// what makes the difference rather than the mutual reachability underneath.
		for eps := 0.02; eps <= 3.0; eps += 0.01 {
			if separatesGroups(h.ExtractDBSCAN(eps), truth, 3, 32) {
				assert.Fail(t, "a single cut separated all three groups", "eps %g", eps)
			}
		}
	})
}

// TestHierarchyScores covers the per-point confidence the condensed tree yields, which is the
// by-product that would answer markers.face_p without a second model.
func TestHierarchyScores(t *testing.T) {
	data, truth := variableDensity(t)

	h, err := HDBSCAN(data, 5, 5, 1, EuclideanDist)
	assert.NoError(t, err)

	labels, probs, outliers := h.Labels(), h.Probabilities(), h.Outliers()

	t.Run("InRange", func(t *testing.T) {
		assert.Len(t, probs, len(data))
		assert.Len(t, outliers, len(data))

		for i := range data {
			assert.GreaterOrEqual(t, probs[i], 0.0)
			assert.LessOrEqual(t, probs[i], 1.0)
			assert.GreaterOrEqual(t, outliers[i], 0.0)
			assert.LessOrEqual(t, outliers[i], 1.0)
		}
	})
	t.Run("NoiseHasNoMembership", func(t *testing.T) {
		for i, l := range labels {
			if l == Noise {
				assert.Zero(t, probs[i])
			}
		}
	})
	t.Run("MembershipAndOutlierDisagree", func(t *testing.T) {
		// Both read the level a point left its cluster at, from opposite ends, so a confident
		// member cannot also score as an outlier.
		for i, l := range labels {
			if l != Noise && probs[i] > 0.99 {
				assert.Less(t, outliers[i], 0.01)
			}
		}
	})
	t.Run("RanksCentralityWithinACluster", func(t *testing.T) {
		// What the score measures: the point nearest a group's center outscores the one farthest
		// from it, in every group and whatever that group's density.
		for group := 1; group <= 3; group++ {
			near, far := extremes(data, labels, truth, group)
			assert.Greater(t, probs[near], probs[far], "group %d ranks its outermost point above its center", group)
		}
	})
	t.Run("TheScaleIsPerClusterRatherThanShared", func(t *testing.T) {
		// Every cluster reaches one, however sparse it is, so the score says where a point sits
		// within its own cluster and not how good that cluster is. Two clusters' values are not
		// comparable, which bears on reading it as a stored confidence.
		for group := 1; group <= 3; group++ {
			_, top := groupRange(probs, labels, truth, group)
			assert.InDelta(t, 1.0, top, 1e-9, "group %d never reaches full membership", group)
		}
	})
}

// TestHierarchyExtractDBSCAN covers the control arm, which cuts the hierarchy at one distance the
// way DBSCAN does, so both can be compared on one code path.
func TestHierarchyExtractDBSCAN(t *testing.T) {
	data := blobs(t)

	h, err := HDBSCAN(data, 5, 5, 1, EuclideanDist)
	assert.NoError(t, err)

	t.Run("ThreeGroupsAtAModerateDistance", func(t *testing.T) {
		assert.Equal(t, 3, h.ExtractDBSCAN(1.0).Count())
	})
	t.Run("NothingBelowTheCoreDistance", func(t *testing.T) {
		assert.Equal(t, len(data), h.ExtractDBSCAN(0.01).NoiseCount())
	})
	t.Run("OneGroupWhenEverythingConnects", func(t *testing.T) {
		assert.Equal(t, 1, h.ExtractDBSCAN(20).Count())
	})
	t.Run("AgreesWithDBSCANOnCorePoints", func(t *testing.T) {
		// Cutting the hierarchy leaves out the border points DBSCAN keeps, so the comparison is
		// over the points that carry a cluster.
		for _, eps := range []float64{0.5, 0.8, 1.2} {
			cut, direct := h.ExtractDBSCAN(eps), dbscanLabels(t, data, 5, eps)

			for i := range cut {
				if h.CoreDist[i] > eps {
					cut[i], direct[i] = Noise, Noise
				}
			}

			assert.True(t, sameClustering(cut, direct), "the cut and DBSCAN disagree at eps %g", eps)
		}
	})
}

// TestMembershipAndOutlierScore covers the two scores at the boundaries the tree produces, where a
// cluster never dies or a point never leaves it.
func TestMembershipAndOutlierScore(t *testing.T) {
	t.Run("Membership", func(t *testing.T) {
		assert.InDelta(t, 0.5, membership(1.0, 0.5), 1e-9)
		assert.InDelta(t, 1.0, membership(1.0, 2.0), 1e-9, "a level beyond the cluster's own is full membership")
		assert.InDelta(t, 1.0, membership(0, 0.5), 1e-9, "a cluster that never dies leaves every point a member")
		assert.InDelta(t, 1.0, membership(1.0, math.Inf(1)), 1e-9, "coincident points never separate")
	})
	t.Run("OutlierScore", func(t *testing.T) {
		assert.InDelta(t, 0.5, outlierScore(1.0, 0.5), 1e-9)
		assert.Zero(t, outlierScore(1.0, 2.0), "a point that outlasts its cluster is no outlier")
		assert.Zero(t, outlierScore(0, 0.5))
		assert.Zero(t, outlierScore(1.0, math.Inf(1)))
	})
}

func TestLambdaOf(t *testing.T) {
	assert.InDelta(t, 0.5, lambdaOf(2), 1e-9)
	assert.True(t, math.IsInf(lambdaOf(0), 1), "coincident points separate at no density")
	assert.True(t, math.IsInf(lambdaOf(-1), 1))
}

// extremes returns the clustered points of one ground-truth group that sit nearest to and farthest
// from its center.
func extremes(data [][]float64, labels Labels, truth []int, group int) (int, int) {
	var (
		center []float64
		n      int
	)

	for i, g := range truth {
		if g != group {
			continue
		}

		if center == nil {
			center = make([]float64, len(data[i]))
		}

		for d := range center {
			center[d] += data[i][d]
		}

		n++
	}

	for d := range center {
		center[d] /= float64(n)
	}

	near, far := -1, -1
	nearest, farthest := math.Inf(1), 0.0

	for i, g := range truth {
		if g != group || labels[i] < 1 {
			continue
		}

		d := EuclideanDist(data[i], center)

		// Independently, since the first point examined is both the nearest and the farthest so
		// far and a single chain would leave one of them unset.
		if d < nearest {
			near, nearest = i, d
		}

		if d >= farthest {
			far, farthest = i, d
		}
	}

	return near, far
}

// groupRange returns the lowest and highest membership among the clustered points of one group.
func groupRange(probs []float64, labels Labels, truth []int, group int) (float64, float64) {
	low, high := math.Inf(1), 0.0

	for i, g := range truth {
		if g != group || labels[i] < 1 {
			continue
		}

		low, high = math.Min(low, probs[i]), math.Max(high, probs[i])
	}

	return low, high
}
