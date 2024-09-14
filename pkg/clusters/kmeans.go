package clusters

import (
	"math"
	"math/rand/v2"
	"sync"

	"gonum.org/v1/gonum/floats"
)

const (
	changesThreshold = 2
)

type kmeansClusterer struct {
	iterations, number int

	// variables keeping count of changes of points' membership every iteration. User as a stopping condition.
	changes, oldchanges, counter, threshold int

	// For online learning only
	alpha     float64
	dimension int

	distance DistFunc

	// slices holding the cluster mapping and sizes. Access is synchronized to avoid read during computation.
	mu   sync.RWMutex
	a, b []int

	// slices holding values of centroids of each clusters
	m, n [][]float64

	// dataset
	d [][]float64
}

// Implementation of k-means++ algorithm with online learning
func KMeans(iterations, clusters int, distance DistFunc) (HardClusterer, error) {
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

	return &kmeansClusterer{
		iterations: iterations,
		number:     clusters,
		distance:   d,
	}, nil
}

func (c *kmeansClusterer) IsOnline() bool {
	return true
}

func (c *kmeansClusterer) WithOnline(o Online) HardClusterer {
	c.alpha = o.Alpha
	c.dimension = o.Dimension

	c.d = make([][]float64, 0, 100)

	c.initializeMeans()

	return c
}

func (c *kmeansClusterer) Learn(data [][]float64) error {
	if len(data) == 0 {
		return errEmptySet
	}

	c.mu.Lock()

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

	c.n = nil

	c.mu.Unlock()

	return nil
}

func (c *kmeansClusterer) Sizes() []int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.b
}

func (c *kmeansClusterer) Guesses() []int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.a
}

func (c *kmeansClusterer) Predict(p []float64) int {
	var (
		l int
		d float64
		m float64 = c.distance(p, c.m[0])
	)

	for i := 1; i < c.number; i++ {
		if d = c.distance(p, c.m[i]); d < m {
			m = d
			l = i
		}
	}

	return l
}

func (c *kmeansClusterer) Online(observations chan []float64, done chan struct{}) chan *HCEvent {
	c.mu.Lock()

	var (
		r    chan *HCEvent = make(chan *HCEvent)
		l, f int           = len(c.m), len(c.m[0])
		h    float64       = 1 - c.alpha
	)

	c.b = make([]int, c.number)

	/* The first step of online learning is adjusting the centroids by finding the one closes to new data point
	 * and modifying it's location using given alpha. Once the client quits sending new data, the actual clusters
	 * are computed and the mutex is unlocked. */

	go func() {
		for {
			select {
			case o := <-observations:
				var (
					k int
					n float64
					m float64 = math.Pow(c.distance(o, c.m[0]), 2)
				)

				for i := 1; i < l; i++ {
					if n = math.Pow(c.distance(o, c.m[i]), 2); n < m {
						m = n
						k = i
					}
				}

				r <- &HCEvent{
					Cluster:     k,
					Observation: o,
				}

				for i := 0; i < f; i++ {
					c.m[k][i] = c.alpha*o[i] + h*c.m[k][i]
				}

				c.d = append(c.d, o)
			case <-done:
				go func() {
					var (
						n    int
						d, m float64
					)

					c.a = make([]int, len(c.d))

					for i := 0; i < len(c.d); i++ {
						m = c.distance(c.d[i], c.m[0])
						n = 0

						for j := 1; j < c.number; j++ {
							if d = c.distance(c.d[i], c.m[j]); d < m {
								m = d
								n = j
							}
						}

						c.a[i] = n + 1
						c.b[n]++
					}

					c.mu.Unlock()
				}()

				return
			}
		}
	}()

	return r
}

// private
func (c *kmeansClusterer) initializeMeansWithData() {
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

func (c *kmeansClusterer) initializeMeans() {
	c.m = make([][]float64, c.number)

	for i := 0; i < c.number; i++ {
		c.m[i] = make([]float64, c.dimension)
		for j := 0; j < c.dimension; j++ {
			c.m[i][j] = 10 * (rand.Float64() - 0.5)
		}
	}
}

func (c *kmeansClusterer) run() {
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

func (c *kmeansClusterer) check() {
	if c.changes == c.oldchanges {
		c.counter++
	}

	c.oldchanges = c.changes
}
