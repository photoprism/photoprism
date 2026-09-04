package alg

import (
	"math"
	"math/rand"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabels(t *testing.T) {
	t.Run("Counts", func(t *testing.T) {
		l := Labels{1, 1, 2, Noise, 2, 2}
		assert.Equal(t, 2, l.Count())
		assert.Equal(t, []int{2, 3}, l.Sizes())
		assert.Equal(t, 1, l.NoiseCount())
	})
	t.Run("AllNoise", func(t *testing.T) {
		l := Labels{Noise, Noise}
		assert.Equal(t, 0, l.Count())
		assert.Empty(t, l.Sizes())
		assert.Equal(t, 2, l.NoiseCount())
	})
	t.Run("Empty", func(t *testing.T) {
		var l Labels
		assert.Equal(t, 0, l.Count())
		assert.Equal(t, 0, l.NoiseCount())
	})
}

func TestInsertSorted(t *testing.T) {
	t.Run("KeepsSmallest", func(t *testing.T) {
		s := []float64{}
		for _, v := range []float64{5, 1, 4, 2, 3} {
			s = insertSorted(s, v, 3)
		}
		assert.Equal(t, []float64{1, 2, 3}, s)
	})
	t.Run("RejectsAboveLimit", func(t *testing.T) {
		s := insertSorted(insertSorted([]float64{}, 1, 2), 2, 2)
		assert.Equal(t, []float64{1, 2}, insertSorted(s, 9, 2))
	})
	t.Run("BelowLimit", func(t *testing.T) {
		assert.Equal(t, []float64{2, 7}, insertSorted([]float64{7}, 2, 4))
	})
}

// TestCoreDistances pins the convention that a point is its own neighbor, which is what makes
// minPts mean the same number of faces here as it does in DBSCAN.
func TestCoreDistances(t *testing.T) {
	// Four points on a line at 0, 1, 2 and 10.
	data := [][]float64{{0}, {1}, {2}, {10}}

	t.Run("CountsThePointItself", func(t *testing.T) {
		// minPts 2 asks for one other point, so the core distance is the nearest neighbor.
		core := coreDistances(data, 2, math.Inf(1), 1, EuclideanDist)
		assert.Equal(t, []float64{1, 1, 1, 8}, core)
	})
	t.Run("SecondNeighbor", func(t *testing.T) {
		core := coreDistances(data, 3, math.Inf(1), 1, EuclideanDist)
		assert.Equal(t, []float64{2, 1, 2, 9}, core)
	})
	t.Run("MinPtsOneIsZero", func(t *testing.T) {
		assert.Equal(t, []float64{0, 0, 0, 0}, coreDistances(data, 1, math.Inf(1), 1, EuclideanDist))
	})
	t.Run("TooFewWithinMaxEps", func(t *testing.T) {
		// The isolated point has no neighbor within 3, so it has no core distance at all.
		core := coreDistances(data, 2, 3, 1, EuclideanDist)
		assert.Equal(t, []float64{1, 1, 1}, core[:3])
		assert.True(t, math.IsInf(core[3], 1))
	})
	t.Run("WorkersDoNotChangeTheResult", func(t *testing.T) {
		wide := blobs(t)
		one := coreDistances(wide, 5, math.Inf(1), 1, EuclideanDist)
		many := coreDistances(wide, 5, math.Inf(1), 8, EuclideanDist)
		assert.Equal(t, one, many)
	})
}

// TestParallelRows covers the chunked dispatch, including the sizes where the last worker's range
// is short or empty. The clustering fixtures all sit below parallelRowsMin, so without this the
// parallel path would never run under test.
func TestParallelRows(t *testing.T) {
	t.Run("CoversEveryIndexOnce", func(t *testing.T) {
		for _, n := range []int{0, 1, 7, 100, parallelRowsMin, parallelRowsMin + 1, 5000} {
			for _, workers := range []int{1, 3, 8, 64} {
				seen := make([]int32, n)

				parallelRows(n, workers, func(i int) { atomic.AddInt32(&seen[i], 1) })

				for i := range seen {
					assert.Equal(t, int32(1), seen[i], "index %d of %d with %d workers", i, n, workers)
				}
			}
		}
	})
	t.Run("MoreWorkersThanRows", func(t *testing.T) {
		seen := make([]int32, 3)
		parallelRows(3, 100, func(i int) { atomic.AddInt32(&seen[i], 1) })
		assert.Equal(t, []int32{1, 1, 1}, seen)
	})
}

// TestCoreDistancesAboveTheParallelThreshold checks that the chunked path agrees with the serial
// one on a set large enough to take it.
func TestCoreDistancesAboveTheParallelThreshold(t *testing.T) {
	//nolint:gosec // a fixed seed is what makes the fixture reproducible.
	r := rand.New(rand.NewSource(5))

	data := gaussianBlob(r, []float64{0, 0, 0}, 1, parallelRowsMin+250)

	assert.Equal(t,
		coreDistances(data, 5, math.Inf(1), 1, EuclideanDist),
		coreDistances(data, 5, math.Inf(1), 8, EuclideanDist))
}
