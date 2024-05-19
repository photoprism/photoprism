package clusters

import (
	"math"
	"math/rand/v2"

	"gonum.org/v1/gonum/floats"
)

type kmeansEstimator struct {
	iterations, number, max int

	// variables keeping count of changes of points' membership every iteration. User as a stopping condition.
	changes, oldchanges, counter, threshold int

	distance DistFunc

	a, b []int

	// slices holding values of centroids of each clusters
	m, n [][]float64

	// dataset
	d [][]float64
}

// KMeansEstimator implements a cluster number estimator using the gap statistic ("Estimating the number of clusters
// in a data set via the gap statistic", Tibshirani et al.) with k-means++ as clustering algorithm.
func KMeansEstimator(iterations, clusters int, distance DistFunc) (Estimator, error) {
	if iterations < 1 {
		return nil, errZeroIterations
	}

	if clusters < 2 {
		return nil, errOneCluster
	}

	var d DistFunc
	{
		if distance != nil {
			d = distance
		} else {
			d = EuclideanDist
		}
	}

	return &kmeansEstimator{
		iterations: iterations,
		max:        clusters,
		distance:   d,
	}, nil
}

func (c *kmeansEstimator) Estimate(data [][]float64) (int, error) {
	if len(data) == 0 {
		return 0, errEmptySet
	}

	var (
		estimated = 0
		size      = len(data)
		bounds    = bounds(data)
		wks       = make([]float64, c.max)
		wkbs      = make([]float64, c.max)
		sk        = make([]float64, c.max)
		one       = make([]float64, c.max)
		bwkbs     = make([]float64, c.max)
	)

	for i := 0; i < c.max; i++ {
		c.number = i + 1

		c.learn(data)

		wks[i] = math.Log(c.wk(c.d, c.m, c.a))

		for j := 0; j < c.max; j++ {
			c.learn(c.buildRandomizedSet(size, bounds))

			bwkbs[j] = math.Log(c.wk(c.d, c.m, c.a))
			one[j] = 1
		}

		wkbs[i] = floats.Sum(bwkbs) / float64(c.max)

		floats.Scale(wkbs[i], one)
		floats.Sub(bwkbs, one)
		floats.Mul(bwkbs, bwkbs)

		sk[i] = math.Sqrt(floats.Sum(bwkbs) / float64(c.max))
	}

	floats.Scale(math.Sqrt(1+(1/float64(c.max))), sk)

	for i := 0; i < c.max-1; i++ {
		if wkbs[i] >= wkbs[i+1]-sk[i+1] {
			estimated = i + 1
			break
		}
	}

	return estimated, nil
}

// private
func (c *kmeansEstimator) learn(data [][]float64) {
	c.d = data

	c.a = make([]int, len(data))
	c.b = make([]int, c.number)

	c.counter = 0
	c.threshold = changesThreshold
	c.changes = 0
	c.oldchanges = 0

	c.initializeMeansWithData()

	for i := 0; i < c.iterations && c.counter != c.threshold; i++ {
		c.run()
		c.check()
	}
}

func (c *kmeansEstimator) initializeMeansWithData() {
	c.m = make([][]float64, c.number)
	c.n = make([][]float64, c.number)

	var (
		k          int
		s, t, l, f float64
		d          []float64 = make([]float64, len(c.d))
	)

	c.m[0] = c.d[rand.IntN(len(c.d)-1)]

	for i := 1; i < c.number; i++ {
		s = 0
		t = 0
		for j := 0; j < len(c.d); j++ {

			l = c.distance(c.m[0], c.d[j])
			for g := 1; g < i; g++ {
				if f = c.distance(c.m[g], c.d[j]); f < l {
					l = f
				}
			}

			d[j] = math.Pow(l, 2)
			s += d[j]
		}

		t = rand.Float64() * s
		k = 0
		for s = d[0]; s < t; s += d[k] {
			k++
		}

		c.m[i] = c.d[k]
	}

	for i := 0; i < c.number; i++ {
		c.n[i] = make([]float64, len(c.m[0]))
	}
}

func (c *kmeansEstimator) run() {
	var (
		l, k, n int = len(c.m[0]), 0, 0
		m, d    float64
	)

	for i := 0; i < c.number; i++ {
		c.b[i] = 0
	}

	for i := 0; i < len(c.d); i++ {
		m = c.distance(c.d[i], c.m[0])
		n = 0

		for j := 1; j < c.number; j++ {
			if d = c.distance(c.d[i], c.m[j]); d < m {
				m = d
				n = j
			}
		}

		k = n + 1

		if c.a[i] != k {
			c.changes++
		}

		c.a[i] = k
		c.b[n]++

		floats.Add(c.n[n], c.d[i])
	}

	for i := 0; i < c.number; i++ {
		floats.Scale(1/float64(c.b[i]), c.n[i])

		for j := 0; j < l; j++ {
			c.m[i][j] = c.n[i][j]
			c.n[i][j] = 0
		}
	}
}

func (c *kmeansEstimator) check() {
	if c.changes == c.oldchanges {
		c.counter++
	}

	c.oldchanges = c.changes
}

func (c *kmeansEstimator) wk(data [][]float64, centroids [][]float64, mapping []int) float64 {
	var (
		l  = float64(2 * len(data[0]))
		wk = make([]float64, len(centroids))
	)

	for i := 0; i < len(mapping); i++ {
		wk[mapping[i]-1] += EuclideanDistSquared(centroids[mapping[i]-1], data[i]) / l
	}

	return floats.Sum(wk)
}

func (c *kmeansEstimator) buildRandomizedSet(size int, bounds []*[2]float64) [][]float64 {
	var (
		l = len(bounds)
		r = make([][]float64, size)
	)

	for i := 0; i < size; i++ {
		r[i] = make([]float64, l)

		for j := 0; j < l; j++ {
			r[i][j] = uniform(bounds[j])
		}
	}

	return r
}
