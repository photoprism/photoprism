package alg

import (
	"math"
	"runtime"
	"sync"
)

// Noise is the label of a point that no cluster contains.
const Noise = -1

// Labels maps each data point to the cluster holding it, numbered from one, with Noise for points
// no cluster took. The numbering matches HardClusterer.Guesses, so both feed the same reporting.
type Labels []int

// Count returns the number of distinct clusters.
func (l Labels) Count() int {
	n := 0

	for _, v := range l {
		if v > n {
			n = v
		}
	}

	return n
}

// Sizes returns the number of points in each cluster, indexed by label minus one.
func (l Labels) Sizes() []int {
	sizes := make([]int, l.Count())

	for _, v := range l {
		if v > 0 {
			sizes[v-1]++
		}
	}

	return sizes
}

// NoiseCount returns the number of points that no cluster took.
func (l Labels) NoiseCount() int {
	n := 0

	for _, v := range l {
		if v < 1 {
			n++
		}
	}

	return n
}

// numCPU returns the worker count to use, resolving zero to the number of processors.
func numCPU(workers int) int {
	if workers > 0 {
		return workers
	}

	return max(1, runtime.NumCPU())
}

// parallelRows runs fn for every index below n, spread over the given number of workers.
//
// Each worker takes one contiguous range rather than one index at a time. The callers run this
// once per point, so handing out indices individually would cost n^2 channel operations and come
// to dominate the distances it is meant to parallelize.
func parallelRows(n, workers int, fn func(i int)) {
	w := min(numCPU(workers), n)

	// Below a few thousand rows the goroutines cost more than the work they take on.
	if w < 2 || n < parallelRowsMin {
		for i := range n {
			fn(i)
		}

		return
	}

	var wg sync.WaitGroup

	size := (n + w - 1) / w

	wg.Add(w)

	for c := range w {
		go func(start int) {
			defer wg.Done()

			for i := start; i < min(start+size, n); i++ {
				fn(i)
			}
		}(c * size)
	}

	wg.Wait()
}

// parallelRowsMin is the row count below which a pass runs on the calling goroutine.
const parallelRowsMin = 2048

// coreDistances returns the distance from every point to its minPts-th nearest neighbor, counting
// the point itself, and Inf where fewer than minPts points lie within maxEps.
//
// Counting the point itself matches the convention DBSCAN in this package uses, where a point is
// its own neighbor, so that minPts means the same number of faces in both.
func coreDistances(data [][]float64, minPts int, maxEps float64, workers int, distance DistFunc) []float64 {
	n := len(data)
	core := make([]float64, n)

	// The point itself already satisfies a core of one, at a distance of zero.
	if minPts < 2 {
		return core
	}

	// How many neighbors other than the point itself the core distance is measured to.
	k := minPts - 1

	parallelRows(n, workers, func(i int) {
		// Ascending, holding at most the k smallest distances seen so far. k is the cluster core
		// size, so an insertion sort over it costs less than a heap.
		nearest := make([]float64, 0, k)

		for j := range n {
			if j == i {
				continue
			}

			d := distance(data[i], data[j])

			if math.IsNaN(d) || d > maxEps {
				continue
			}

			if len(nearest) == k && d >= nearest[k-1] {
				continue
			}

			nearest = insertSorted(nearest, d, k)
		}

		if len(nearest) < k {
			core[i] = math.Inf(1)
		} else {
			core[i] = nearest[k-1]
		}
	})

	return core
}

// insertSorted places v in the ascending slice, keeping it no longer than the given limit.
func insertSorted(s []float64, v float64, limit int) []float64 {
	pos := len(s)

	for i, x := range s {
		if v < x {
			pos = i
			break
		}
	}

	if len(s) < limit {
		s = append(s, 0)
	} else if pos == limit {
		return s
	}

	copy(s[pos+1:], s[pos:])
	s[pos] = v

	return s
}
