package alg

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOPTICS(t *testing.T) {
	t.Run("EmptySet", func(t *testing.T) {
		_, err := OPTICS(nil, 5, 1, 1, EuclideanDist)
		assert.ErrorIs(t, err, errEmptySet)
	})
	t.Run("RaggedData", func(t *testing.T) {
		_, err := OPTICS([][]float64{{1, 2}, {3}}, 2, 1, 1, EuclideanDist)
		assert.ErrorIs(t, err, errRaggedData)
	})
	t.Run("ZeroMinPts", func(t *testing.T) {
		_, err := OPTICS([][]float64{{1}, {2}}, 0, 1, 1, EuclideanDist)
		assert.ErrorIs(t, err, errZeroMinpts)
	})
	t.Run("NegativeWorkers", func(t *testing.T) {
		_, err := OPTICS([][]float64{{1}, {2}}, 2, 1, -1, EuclideanDist)
		assert.ErrorIs(t, err, errZeroWorkers)
	})
	t.Run("ZeroMaxEps", func(t *testing.T) {
		_, err := OPTICS([][]float64{{1}, {2}}, 2, 0, 1, EuclideanDist)
		assert.ErrorIs(t, err, errZeroEpsilon)
	})
	t.Run("DefaultDistance", func(t *testing.T) {
		o, err := OPTICS([][]float64{{0}, {1}, {2}}, 2, math.Inf(1), 1, nil)
		assert.NoError(t, err)
		assert.Len(t, o.Order, 3)
	})
	t.Run("OrdersEveryPointOnce", func(t *testing.T) {
		data := blobs(t)
		o, err := OPTICS(data, 5, math.Inf(1), 4, EuclideanDist)
		assert.NoError(t, err)
		assert.Len(t, o.Order, len(data))

		seen := make(map[int]bool, len(data))
		for _, p := range o.Order {
			assert.False(t, seen[p], "point %d appears twice in the ordering", p)
			seen[p] = true
		}
		assert.Len(t, seen, len(data))
	})
	t.Run("FirstPointIsUnreachable", func(t *testing.T) {
		o, err := OPTICS(blobs(t), 5, math.Inf(1), 1, EuclideanDist)
		assert.NoError(t, err)
		assert.True(t, math.IsInf(o.Plot()[0], 1), "the point a run starts from has nothing to be reached from")
	})
	t.Run("ReachabilityNeverBelowTheCoreDistance", func(t *testing.T) {
		data := blobs(t)
		o, err := OPTICS(data, 5, math.Inf(1), 1, EuclideanDist)
		assert.NoError(t, err)

		// Every reachability is the core distance of the point it was measured from, or the
		// plain distance where that is larger. Neither can fall below the smaller core.
		for i, p := range o.Order {
			if i == 0 || math.IsInf(o.Reachability[p], 1) {
				continue
			}
			pred := o.Predecessor[p]
			assert.GreaterOrEqual(t, o.Reachability[p], o.CoreDist[pred]-1e-12)
		}
	})
	t.Run("WorkersDoNotChangeTheOrdering", func(t *testing.T) {
		data := blobs(t)
		one, err := OPTICS(data, 5, math.Inf(1), 1, EuclideanDist)
		assert.NoError(t, err)
		many, err := OPTICS(data, 5, math.Inf(1), 8, EuclideanDist)
		assert.NoError(t, err)
		assert.Equal(t, one.Order, many.Order)
		assert.Equal(t, one.Reachability, many.Reachability)
	})
	t.Run("MaxEpsLeavesDistantPointsUnreachable", func(t *testing.T) {
		data := blobs(t)
		// Well below the 10 that separates the groups, so no reachability crosses between them.
		o, err := OPTICS(data, 5, 2, 1, EuclideanDist)
		assert.NoError(t, err)

		unreachable := 0
		for _, r := range o.Reachability {
			if math.IsInf(r, 1) {
				unreachable++
			}
		}
		assert.Equal(t, 3, unreachable, "one point per group starts a run of its own")
	})
}

// TestOrderingExtractDBSCAN is the correctness anchor for the ordering: an extraction at eps has to
// reproduce a DBSCAN run at the same eps on the core points, since OPTICS contains DBSCAN as a
// special case.
func TestOrderingExtractDBSCAN(t *testing.T) {
	data := blobs(t)

	o, err := OPTICS(data, 5, math.Inf(1), 1, EuclideanDist)
	assert.NoError(t, err)

	t.Run("AgreesWithDBSCAN", func(t *testing.T) {
		// Core points only. This extraction takes a border point in the order the ordering reached
		// it, while this package's DBSCAN attaches one only where the cores around it agree, so the
		// two are comparable on the points a core distance already places.
		for _, eps := range []float64{0.5, 0.8, 1.2, 2.0} {
			extracted, direct := o.ExtractDBSCAN(eps), dbscanLabels(t, data, 5, eps)

			for i := range extracted {
				if o.CoreDist[i] > eps {
					extracted[i], direct[i] = Noise, Noise
				}
			}

			assert.True(t, sameClustering(extracted, direct),
				"extraction and DBSCAN disagree on the core points at eps %g", eps)
		}
	})
	t.Run("ThreeGroupsAtAModerateDistance", func(t *testing.T) {
		assert.Equal(t, 3, o.ExtractDBSCAN(1.0).Count())
	})
	t.Run("NothingBelowTheCoreDistance", func(t *testing.T) {
		assert.Equal(t, len(data), o.ExtractDBSCAN(0.01).NoiseCount())
	})
	t.Run("OneGroupWhenEverythingConnects", func(t *testing.T) {
		assert.Equal(t, 1, o.ExtractDBSCAN(20).Count())
	})
}

// TestOrderingExtractXi covers the property the whole evaluation rests on: valleys of different
// depth are separated in one pass, where no single link distance separates them at all.
func TestOrderingExtractXi(t *testing.T) {
	data, truth := variableDensity(t)

	o, err := OPTICS(data, 5, math.Inf(1), 4, EuclideanDist)
	assert.NoError(t, err)

	t.Run("NoLinkDistanceSeparatesTheThreeGroups", func(t *testing.T) {
		// The premise of the fixture, asserted rather than assumed: every distance either
		// fragments the sparse group or chains the dense pair into one cluster.
		for eps := 0.02; eps <= 3.0; eps += 0.01 {
			if separatesGroups(o.ExtractDBSCAN(eps), truth, 3, 32) {
				assert.Fail(t, "a single link distance separated all three groups", "eps %g", eps)
			}
		}
	})
	t.Run("SeparatesTheDensePairFromTheSparseGroup", func(t *testing.T) {
		// What the evaluation turns on: one pass returns all three, which the loop above has
		// just shown no link distance does.
		assert.True(t, separatesGroups(o.ExtractXi(0.5, 5), truth, 3, 32))
	})
	t.Run("ASmallXiOverSplits", func(t *testing.T) {
		// The knob is not free of tuning. At the value scikit-learn defaults to, every shallow
		// dip within a group reads as a wall and most of the set is left unclustered.
		small := o.ExtractXi(0.05, 5)
		assert.Greater(t, small.Count(), 3)
		assert.Greater(t, small.NoiseCount(), len(data)/2)
	})
	t.Run("RejectsAnInvalidXi", func(t *testing.T) {
		assert.Equal(t, len(data), o.ExtractXi(0, 5).NoiseCount())
		assert.Equal(t, len(data), o.ExtractXi(1, 5).NoiseCount())
		assert.Equal(t, len(data), o.ExtractXi(-1, 5).NoiseCount())
	})
	t.Run("MinSizeNeverBelowMinPts", func(t *testing.T) {
		// A size below the core cannot be honored, since a valley that small holds no core point.
		assert.True(t, sameClustering(o.ExtractXi(0.5, 1), o.ExtractXi(0.5, 5)))
	})
	t.Run("LargeMinSizeDropsEveryCluster", func(t *testing.T) {
		assert.Equal(t, 0, o.ExtractXi(0.5, len(data)+1).Count())
	})
	t.Run("EveryPointLabelledOnce", func(t *testing.T) {
		labels := o.ExtractXi(0.5, 5)
		assert.Len(t, labels, len(data))
		for _, l := range labels {
			assert.True(t, l == Noise || l >= 1)
		}
	})
}

// TestOrderingPlot pins that the plot is in processing order rather than point order, which is
// what the extractions read and the only order the valleys appear in.
func TestOrderingPlot(t *testing.T) {
	data := blobs(t)

	o, err := OPTICS(data, 5, math.Inf(1), 1, EuclideanDist)
	assert.NoError(t, err)

	plot := o.Plot()
	assert.Len(t, plot, len(o.Order))

	for i, p := range o.Order {
		assert.Equal(t, o.Reachability[p], plot[i])
	}
}
