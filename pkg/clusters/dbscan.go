package clusters

import (
	"sync"
)

type dbscanClusterer struct {
	minpts, workers int
	eps             float64

	distance DistFunc

	// slices holding the cluster mapping and sizes. Access is synchronized to avoid read during computation.
	mu sync.RWMutex
	// groups for dateset
	a []int
	b []int

	// variables used for concurrent computation of nearest neighbours
	// dataset len
	l int
	// worker number
	s int
	// work number for per worker
	f int
	j chan *rangeJob
	m *sync.Mutex
	w *sync.WaitGroup
	// current point near
	r *[]int
	// current point
	p []float64

	// visited points
	v []bool

	// dataset
	d [][]float64
}

// Implementation of DBSCAN algorithm with concurrent nearest neighbour computation. The number of goroutines acting concurrently
// is controlled via workers argument. Passing 0 will result in this number being chosen arbitrarily.
func DBSCAN(minpts int, eps float64, workers int, distance DistFunc) (HardClusterer, error) {
	if minpts < 1 {
		return nil, errZeroMinpts
	}

	if workers < 0 {
		return nil, errZeroWorkers
	}

	if eps <= 0 {
		return nil, errZeroEpsilon
	}

	var d DistFunc
	{
		if distance != nil {
			d = distance
		} else {
			d = EuclideanDist
		}
	}

	return &dbscanClusterer{
		minpts:   minpts,
		workers:  workers,
		eps:      eps,
		distance: d,
	}, nil
}

func (c *dbscanClusterer) IsOnline() bool {
	return false
}

func (c *dbscanClusterer) WithOnline(o Online) HardClusterer {
	return c
}

func (c *dbscanClusterer) Learn(data [][]float64) error {
	if len(data) == 0 {
		return errEmptySet
	}

	c.mu.Lock()

	c.l = len(data)
	c.s = c.numWorkers()
	c.f = c.l / c.s

	c.d = data

	c.v = make([]bool, c.l)

	c.a = make([]int, c.l)
	c.b = make([]int, 0)

	c.startNearestWorkers()

	c.run()

	c.endNearestWorkers()

	c.v = nil
	c.p = nil
	c.r = nil

	c.mu.Unlock()

	return nil
}

func (c *dbscanClusterer) Sizes() []int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.b
}

func (c *dbscanClusterer) Guesses() []int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.a
}

func (c *dbscanClusterer) Predict(p []float64) int {
	var (
		l int
		d float64
		m float64 = c.distance(p, c.d[0])
	)

	for i := 1; i < len(c.d); i++ {
		if d = c.distance(p, c.d[i]); d < m {
			m = d
			l = i
		}
	}

	return c.a[l]
}

func (c *dbscanClusterer) Online(observations chan []float64, done chan struct{}) chan *HCEvent {
	return nil
}

// private
func (c *dbscanClusterer) run() {
	var (
		n, m, l, k = 1, 0, 0, 0
		ns, nss    = make([]int, 0), make([]int, 0)
	)

	for i := 0; i < c.l; i++ {
		if c.v[i] {
			continue
		}

		c.v[i] = true

		c.nearest(i, &l, &ns)

		if l < c.minpts {
			c.a[i] = -1
		} else {
			c.a[i] = n

			c.b = append(c.b, 0)
			c.b[m]++

			for j := 0; j < l; j++ {
				if !c.v[ns[j]] {
					c.v[ns[j]] = true

					c.nearest(ns[j], &k, &nss)

					if k >= c.minpts {
						l += k
						ns = append(ns, nss...)
					}
				}

				if c.a[ns[j]] == 0 {
					c.a[ns[j]] = n
					c.b[m]++
				}
			}

			n++
			m++
		}
	}
}

/* Divide work among c.s workers, where c.s is determined
 * by the size of the data. This is based on an assumption that neighbour points of p
 * are located in relatively small subsection of the input data, so the dataset can be scanned
 * concurrently without blocking a big number of goroutines trying to write to r */
func (c *dbscanClusterer) nearest(p int, l *int, r *[]int) {
	var b int

	*r = (*r)[:0]

	c.p = c.d[p]
	c.r = r

	for i := 0; i < c.l; i += c.f {
		if c.l-i <= c.f {
			b = c.l
		} else {
			b = i + c.f
		}

		c.w.Add(1)
		c.j <- &rangeJob{
			a: i,
			b: b,
		}
	}

	c.w.Wait()

	*l = len(*r)
}

func (c *dbscanClusterer) startNearestWorkers() {
	c.j = make(chan *rangeJob, c.l)

	c.m = &sync.Mutex{}
	c.w = &sync.WaitGroup{}

	for i := 0; i < c.s; i++ {
		go c.nearestWorker()
	}
}

func (c *dbscanClusterer) endNearestWorkers() {
	close(c.j)

	c.j = nil

	c.m = nil
	c.w = nil
}

func (c *dbscanClusterer) nearestWorker() {
	for j := range c.j {
		for i := j.a; i < j.b; i++ {
			if c.distance(c.p, c.d[i]) < c.eps {
				c.m.Lock()
				*c.r = append(*c.r, i)
				c.m.Unlock()
			}
		}

		c.w.Done()
	}
}

func (c *dbscanClusterer) numWorkers() int {
	var b int

	if c.l < 1000 {
		b = 1
	} else if c.l < 10000 {
		b = 10
	} else if c.l < 100000 {
		b = 100
	} else {
		b = 1000
	}

	if c.workers == 0 {
		return b
	}

	if c.workers < b {
		return c.workers
	}

	return b

}
