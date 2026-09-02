package alg

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestDBSCANCluster(t *testing.T) {
	tests := []struct {
		MinPts   int
		Eps      float64
		Points   [][]float64
		Expected []int
	}{
		{
			MinPts:   1,
			Eps:      1,
			Points:   [][]float64{{1}},
			Expected: []int{1},
		},
		{
			MinPts:   1,
			Eps:      1,
			Points:   [][]float64{{1}, {1.5}},
			Expected: []int{1, 1},
		},
		{
			MinPts:   1,
			Eps:      1,
			Points:   [][]float64{{1}, {1}},
			Expected: []int{1, 1},
		},
		{
			MinPts:   1,
			Eps:      1,
			Points:   [][]float64{{1}, {1}, {1}},
			Expected: []int{1, 1, 1},
		},
		{
			MinPts:   1,
			Eps:      1,
			Points:   [][]float64{{1}, {1.5}, {2}},
			Expected: []int{1, 1, 1},
		},
		{
			MinPts:   1,
			Eps:      1,
			Points:   [][]float64{{1}, {1.5}, {3}},
			Expected: []int{1, 1, 2},
		},
		{
			MinPts:   2,
			Eps:      1,
			Points:   [][]float64{{1}, {3}},
			Expected: []int{-1, -1},
		},
	}
	for _, test := range tests {
		c, e := DBSCAN(test.MinPts, test.Eps, 0, EuclideanDist)
		if e != nil {
			t.Errorf("Error initializing kmeans clusterer: %s\n", e.Error())
		}

		if e = c.Learn(test.Points); e != nil {
			t.Errorf("Error learning data: %s\n", e.Error())
		}

		if !reflect.DeepEqual(c.Guesses(), test.Expected) {
			t.Errorf("guesses does not match: %d vs %d\n", c.Guesses(), test.Expected)
		}
	}
}

func TestPartitionSize(t *testing.T) {
	tests := []struct {
		Points   int
		Workers  int
		Expected int
	}{
		{Points: 1000, Workers: 10, Expected: 100},
		{Points: 10, Workers: 2, Expected: 5},
		{Points: 5, Workers: 1, Expected: 5},
		{Points: 3, Workers: 64, Expected: 1},
		{Points: 1, Workers: 1, Expected: 1},
		{Points: 0, Workers: 1, Expected: 1},
		{Points: 4, Workers: 0, Expected: 4},
		{Points: 4, Workers: -3, Expected: 4},
	}
	for _, test := range tests {
		if got := partitionSize(test.Points, test.Workers); got != test.Expected {
			t.Errorf("partitionSize(%d, %d) = %d, want %d", test.Points, test.Workers, got, test.Expected)
		}
	}
}

func TestDBSCANMoreWorkersThanPoints(t *testing.T) {
	// Requesting far more workers than data points must still terminate and
	// cluster correctly; the partition size is floored at 1.
	c, err := DBSCAN(1, 1, 64, EuclideanDist)
	if err != nil {
		t.Fatalf("unexpected constructor error: %s", err)
	}

	points := [][]float64{{1}, {1.5}, {5}}

	if err = c.Learn(points); err != nil {
		t.Fatalf("unexpected learn error: %s", err)
	}

	if expected := []int{1, 1, 2}; !reflect.DeepEqual(c.Guesses(), expected) {
		t.Errorf("guesses do not match: %d vs %d", c.Guesses(), expected)
	}
}

func TestDBSCANRaggedData(t *testing.T) {
	// Points of different widths cannot be compared, and the distance runs in a worker
	// goroutine whose panic no caller can recover, so Learn must reject them up front.
	c, err := DBSCAN(1, 1, 0, EuclideanDist)
	if err != nil {
		t.Fatalf("unexpected constructor error: %s", err)
	}

	if err = c.Learn([][]float64{{1, 1}, {1}}); err != errRaggedData {
		t.Errorf("expected errRaggedData, got %v", err)
	}
}

func TestDBSCANPredict(t *testing.T) {
	c, err := DBSCAN(1, 1, 0, EuclideanDist)
	if err != nil {
		t.Fatalf("unexpected constructor error: %s", err)
	}
	t.Run("Untrained", func(t *testing.T) {
		if n := c.Predict([]float64{1}); n != -1 {
			t.Errorf("expected -1, got %d", n)
		}
	})
	t.Run("Success", func(t *testing.T) {
		if err = c.Learn([][]float64{{1}, {1.5}, {5}}); err != nil {
			t.Fatalf("unexpected learn error: %s", err)
		}
		if n := c.Predict([]float64{1.2}); n != 1 {
			t.Errorf("expected cluster 1, got %d", n)
		}
	})
	t.Run("WrongDimensions", func(t *testing.T) {
		if n := c.Predict([]float64{1, 2}); n != -1 {
			t.Errorf("expected -1, got %d", n)
		}
	})
}

func TestDBSCANWithProgress(t *testing.T) {
	progress := make([][2]int, 0)

	clusterer, err := DBSCANWithProgress(1, 1, 0, EuclideanDist, time.Second, func(done, total int) {
		progress = append(progress, [2]int{done, total})
	})
	if err != nil {
		t.Fatalf("unexpected constructor error: %s", err)
	}

	c, ok := clusterer.(*dbscanClusterer)
	if !ok {
		t.Fatalf("unexpected clusterer type %T", clusterer)
	}

	current := time.Unix(0, 0)
	c.now = func() time.Time {
		value := current
		current = current.Add(600 * time.Millisecond)
		return value
	}

	points := [][]float64{{1}, {1}, {1}, {1}, {1}}

	if err = c.Learn(points); err != nil {
		t.Fatalf("unexpected learn error: %s", err)
	}

	if len(progress) == 0 {
		t.Fatal("expected at least one progress update")
	}

	// Rising and inside the dataset, rather than positive: finding the cores and walking them are two
	// passes over one scale, and a report from the very start of either is legitimate. What must not
	// happen is the count going backwards, which reads as a restarted job.
	last := -1

	for _, entry := range progress {
		if entry[0] < last {
			t.Fatalf("expected a rising processed count, got %d after %d", entry[0], last)
		}
		if entry[0] < 0 || entry[0] >= len(points) {
			t.Fatalf("expected a processed count inside the dataset, got %d", entry[0])
		}
		if entry[1] != len(points) {
			t.Fatalf("expected total %d, got %d", len(points), entry[1])
		}

		last = entry[0]
	}
}

// borderTestData returns two dense groups of five, far enough apart to stay separate, and a point
// between them that has too few neighbors to be a core of its own.
//
// The spacing is what makes the case: of each group only the point nearest the middle is within eps
// of it, so the middle one sees itself plus one point per group present - two with a single group,
// three with both - and cannot reach minpts either way.
func borderTestData() (left, right [][]float64, middle []float64) {
	left = [][]float64{{0, 0}, {0, 0.02}, {0, 0.04}, {0, 0.06}, {0, 0.15}}
	right = [][]float64{{0, 2.05}, {0, 2.16}, {0, 2.18}, {0, 2.20}, {0, 2.22}}

	return left, right, []float64{0, 1.1}
}

// partitionOf returns which points share a cluster, so two labelings can be compared without
// depending on the numbers each assigned.
func partitionOf(guesses []int) map[int][]int {
	groups := make(map[int][]int)

	for i, g := range guesses {
		if g > 0 {
			groups[g] = append(groups[g], i)
		}
	}

	out := make(map[int][]int, len(groups))

	for _, members := range groups {
		out[members[0]] = members
	}

	return out
}

func TestDBSCANBorderPoints(t *testing.T) {
	left, right, middle := borderTestData()

	t.Run("AttachedToTheOnlyClusterThatReachesIt", func(t *testing.T) {
		points := append(append([][]float64{}, left...), middle)

		c, err := DBSCAN(5, 1, 0, EuclideanDist)
		if err != nil {
			t.Fatalf("unexpected constructor error: %s", err)
		}
		if err = c.Learn(points); err != nil {
			t.Fatalf("unexpected learn error: %s", err)
		}

		if expected := []int{1, 1, 1, 1, 1, 1}; !reflect.DeepEqual(c.Guesses(), expected) {
			t.Errorf("guesses do not match: %d vs %d", c.Guesses(), expected)
		}
		if sizes := c.Sizes(); !reflect.DeepEqual(sizes, []int{6}) {
			t.Errorf("expected the attached point to be counted, got %d", sizes)
		}
	})
	t.Run("NoiseWhenTwoClustersReachIt", func(t *testing.T) {
		points := append(append([][]float64{}, left...), middle)
		points = append(points, right...)

		c, err := DBSCAN(5, 1, 0, EuclideanDist)
		if err != nil {
			t.Fatalf("unexpected constructor error: %s", err)
		}
		if err = c.Learn(points); err != nil {
			t.Fatalf("unexpected learn error: %s", err)
		}

		expected := []int{1, 1, 1, 1, 1, -1, 2, 2, 2, 2, 2}

		if !reflect.DeepEqual(c.Guesses(), expected) {
			t.Errorf("guesses do not match: %d vs %d", c.Guesses(), expected)
		}
		if sizes := c.Sizes(); !reflect.DeepEqual(sizes, []int{5, 5}) {
			t.Errorf("expected an ambiguous point to join neither, got %d", sizes)
		}
	})
	t.Run("DoNotExtendTheCluster", func(t *testing.T) {
		// Reachable from the attached point but from no core, so nothing carries the cluster to it.
		points := append(append([][]float64{}, left...), middle, []float64{0, 1.9})

		c, err := DBSCAN(5, 1, 0, EuclideanDist)
		if err != nil {
			t.Fatalf("unexpected constructor error: %s", err)
		}
		if err = c.Learn(points); err != nil {
			t.Fatalf("unexpected learn error: %s", err)
		}

		if expected := []int{1, 1, 1, 1, 1, 1, -1}; !reflect.DeepEqual(c.Guesses(), expected) {
			t.Errorf("guesses do not match: %d vs %d", c.Guesses(), expected)
		}
	})
}

// TestDBSCANIndependentOfPointOrder pins the property the border rule above is there for: the same
// points presented in a different order have to produce the same clusters.
//
// Many permutations rather than one, because a single one only has about even odds of disturbing a
// given order-dependent implementation - enough to pass while pinning very little.
func TestDBSCANIndependentOfPointOrder(t *testing.T) {
	left, right, middle := borderTestData()

	points := append(append([][]float64{}, left...), middle)
	points = append(points, right...)

	first, err := DBSCAN(5, 1, 0, EuclideanDist)
	if err != nil {
		t.Fatalf("unexpected constructor error: %s", err)
	}
	if err = first.Learn(points); err != nil {
		t.Fatalf("unexpected learn error: %s", err)
	}

	partition := partitionOf(first.Guesses())

	// Guards the comparison below, which two empty partitions would also satisfy.
	if len(partition) != 2 {
		t.Fatalf("expected two clusters to compare, got %d", len(partition))
	}

	//nolint:gosec // a fixed seed is what makes the permutations reproducible.
	r := rand.New(rand.NewSource(1))

	for k := 0; k < 20; k++ {
		order := r.Perm(len(points))

		shuffled := make([][]float64, len(points))

		for i, p := range order {
			shuffled[i] = points[p]
		}

		second, err := DBSCAN(5, 1, 0, EuclideanDist)
		if err != nil {
			t.Fatalf("unexpected constructor error: %s", err)
		}
		if err = second.Learn(shuffled); err != nil {
			t.Fatalf("unexpected learn error: %s", err)
		}

		// Scored back in the original index space, so the permutation itself cannot read as a change.
		back := make([]int, len(points))

		for i, p := range order {
			back[p] = second.Guesses()[i]
		}

		if !reflect.DeepEqual(partition, partitionOf(back)) {
			t.Fatalf("permutation %d changed the clusters: %d vs %d", k, first.Guesses(), back)
		}
	}
}

// TestDBSCANParallelNeighborsAgree covers the same property on the concurrent neighbor search, which
// numWorkers() reserves for a thousand points and up.
//
// Worth its own case because that path fills each neighbor list from several goroutines under a
// mutex, so the list comes back in an order that varies between runs - the one input to clustering
// that the caller cannot control. The clusterer is built by hand rather than through Learn, which
// would derive the worker count from the fixture size and take the single-threaded path.
func TestDBSCANParallelNeighborsAgree(t *testing.T) {
	left, right, middle := borderTestData()

	points := append(append([][]float64{}, left...), middle)
	points = append(points, right...)

	learn := func(data [][]float64) []int {
		c := &dbscanClusterer{minpts: 5, eps: 1, distance: EuclideanDist}
		c.l = len(data)
		c.s = 8
		c.f = partitionSize(c.l, c.s)
		c.d = data
		c.a = make([]int, c.l)
		c.b = make([]int, 0)

		c.startNearestWorkers()
		c.run()
		c.endNearestWorkers()

		return c.a
	}

	expected := partitionOf(learn(points))

	if len(expected) != 2 {
		t.Fatalf("expected two clusters to compare, got %d", len(expected))
	}

	//nolint:gosec // a fixed seed is what makes the permutations reproducible.
	r := rand.New(rand.NewSource(1))

	// Repeated on the same input for run-to-run stability, and on permuted input so that the property
	// itself is covered here rather than only on the single-threaded path.
	for k := 0; k < 8; k++ {
		if got := partitionOf(learn(points)); !reflect.DeepEqual(expected, got) {
			t.Fatalf("run %d disagreed with the first: %v vs %v", k, got, expected)
		}

		order := r.Perm(len(points))

		shuffled := make([][]float64, len(points))

		for i, p := range order {
			shuffled[i] = points[p]
		}

		back := make([]int, len(points))

		for i, p := range order {
			back[p] = learn(shuffled)[i]
		}

		if got := partitionOf(back); !reflect.DeepEqual(expected, got) {
			t.Fatalf("permutation %d disagreed with the first: %v vs %v", k, got, expected)
		}
	}
}

func TestDBSCANCoreFlags(t *testing.T) {
	left, _, middle := borderTestData()

	data := append(append([][]float64{}, left...), middle)

	c := &dbscanClusterer{minpts: 5, eps: 1, distance: EuclideanDist}
	c.l = len(data)
	c.s = c.numWorkers()
	c.f = partitionSize(c.l, c.s)
	c.d = data

	c.startNearestWorkers()
	core := c.coreFlags()
	c.endNearestWorkers()

	if expected := []bool{true, true, true, true, true, false}; !reflect.DeepEqual(core, expected) {
		t.Errorf("core flags do not match: %v vs %v", core, expected)
	}
}
