//nolint:revive,staticcheck,gocritic,gosec // clustering algorithms keep legacy style for clarity
package alg

import (
	"sync"
	"time"
)

type dbscanClusterer struct {
	minpts, workers int
	eps             float64

	distance DistFunc
	now      func() time.Time
	logAfter time.Duration
	logf     func(done, total int)

	// slices holding the cluster mapping and sizes. Access is synchronized to avoid read during computation.
	mu sync.RWMutex
	// groups for dateset
	a []int
	b []int

	// variables used for concurrent computation of nearest neighbors
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

	// dataset
	d [][]float64

	loggedAt time.Time
}

// DBSCAN implements density-based clustering with concurrent nearest neighbor computation.
// The number of goroutines is controlled via workers (0 picks a default).
func DBSCAN(minpts int, eps float64, workers int, distance DistFunc) (HardClusterer, error) {
	return newDBSCANClusterer(minpts, eps, workers, distance, 0, nil)
}

// DBSCANWithProgress implements DBSCAN with optional time-based progress reporting.
func DBSCANWithProgress(minpts int, eps float64, workers int, distance DistFunc, interval time.Duration, progressf func(done, total int)) (HardClusterer, error) {
	return newDBSCANClusterer(minpts, eps, workers, distance, interval, progressf)
}

// newDBSCANClusterer validates the options and creates a DBSCAN clusterer instance.
func newDBSCANClusterer(minpts int, eps float64, workers int, distance DistFunc, interval time.Duration, progressf func(done, total int)) (HardClusterer, error) {
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
		now:      time.Now,
		logAfter: interval,
		logf:     progressf,
	}, nil
}

func (c *dbscanClusterer) IsOnline() bool {
	return false
}

func (c *dbscanClusterer) WithOnline(o Online) HardClusterer {
	return c
}

func (c *dbscanClusterer) Learn(data [][]float64) error {
	if _, err := dataDims(data); err != nil {
		return err
	}

	c.mu.Lock()

	c.l = len(data)
	c.s = c.numWorkers()
	c.f = partitionSize(c.l, c.s)

	c.d = data

	c.a = make([]int, c.l)
	c.b = make([]int, 0)
	c.loggedAt = time.Time{}

	c.startNearestWorkers()

	c.run()

	c.endNearestWorkers()

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
	// Without training data, or for an observation of a different width, there is no
	// cluster to assign, which this algorithm already labels as noise.
	if len(c.d) == 0 || len(p) != len(c.d[0]) {
		return -1
	}

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

// run assigns every point to a cluster or to noise.
//
// Clusters are the connected components of the core points, so which points share one depends on the
// point set rather than on the order it arrives in. Every other point is attached afterwards, and
// only where the cores around it agree on a single cluster: one that two clusters can both reach
// stays noise rather than joining whichever was walked first.
func (c *dbscanClusterer) run() {
	core := c.coreFlags()

	// The cluster a non-core point may join: 0 while none has been seen, -1 once two have.
	border := make([]int, c.l)

	var (
		n     = 1
		l     int
		ns    = make([]int, 0)
		queue = make([]int, 0)
	)

	for i := 0; i < c.l; i++ {
		// The second half of one scale shared with coreFlags, so the two passes report a progress
		// that only ever rises. Reporting each pass against c.l would restart the count midway.
		c.logProgress((c.l + i) / 2)

		if !core[i] || c.a[i] != 0 {
			continue
		}

		c.a[i] = n
		c.b = append(c.b, 1)

		queue = append(queue[:0], i)

		for len(queue) > 0 {
			p := queue[len(queue)-1]
			queue = queue[:len(queue)-1]

			c.nearest(p, &l, &ns)

			for j := 0; j < l; j++ {
				q := ns[j]

				// Only a core extends a cluster, so a point attached below never widens the reach.
				if core[q] {
					if c.a[q] == 0 {
						c.a[q] = n
						c.b[n-1]++
						queue = append(queue, q)
					}

					continue
				}

				// Recorded from the core's side, which is the same set of pairs a symmetric distance
				// would report from the other. An asymmetric DistFunc would make the two differ.
				switch border[q] {
				case 0:
					border[q] = n
				case n, -1:
				default:
					border[q] = -1
				}
			}
		}

		n++
	}

	for i := 0; i < c.l; i++ {
		if core[i] {
			continue
		}

		if b := border[i]; b > 0 {
			c.a[i] = b
			c.b[b-1]++
		} else {
			c.a[i] = -1
		}
	}
}

// coreFlags reports which points hold at least minpts neighbors within eps, the only ones that may
// form a cluster or extend one.
//
// This is a full neighbor scan, so it costs about as much as the pass that follows it.
func (c *dbscanClusterer) coreFlags() []bool {
	var (
		l  int
		ns = make([]int, 0)
	)

	core := make([]bool, c.l)

	for i := 0; i < c.l; i++ {
		c.logProgress(i / 2)
		c.nearest(i, &l, &ns)

		core[i] = l >= c.minpts
	}

	return core
}

// logProgress emits an optional progress update when the reporting interval has elapsed.
func (c *dbscanClusterer) logProgress(done int) {
	if c.logf == nil || c.logAfter <= 0 {
		return
	}

	now := c.now
	if now == nil {
		now = time.Now
	}

	current := now()
	if c.loggedAt.IsZero() {
		c.loggedAt = current
		return
	}

	if current.Sub(c.loggedAt) >= c.logAfter {
		c.logf(done, c.l)
		c.loggedAt = current
	}
}

/* Divide work among c.s workers, where c.s is determined
 * by the size of the data. This is based on an assumption that neighbor points of p
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

// partitionSize returns the per-worker scan range size, floored at 1 so the
// nearest() dispatch loop (which advances the start index by this value) always
// makes progress and terminates, even if the worker count exceeds the number of
// data points. The size-based numWorkers buckets keep points >= workers today,
// so the floor is a defensive guard against future tuning of those buckets.
func partitionSize(points, workers int) int {
	if workers < 1 {
		workers = 1
	}

	return max(1, points/workers)
}

func (c *dbscanClusterer) numWorkers() int {
	var b int

	switch {
	case c.l < 1000:
		b = 1
	case c.l < 10000:
		b = 10
	case c.l < 100000:
		b = 100
	default:
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
