package alg

import (
	"math"
	"math/rand"
	"testing"
)

// gaussianBlob returns n points scattered around a center with the given standard deviation.
func gaussianBlob(r *rand.Rand, center []float64, sd float64, n int) [][]float64 {
	points := make([][]float64, n)

	for i := range n {
		p := make([]float64, len(center))

		for d := range center {
			p[d] = center[d] + r.NormFloat64()*sd
		}

		points[i] = p
	}

	return points
}

// blobs returns three well separated groups of equal density, which every algorithm here has to
// agree on. It is the control the harder sets are read against.
func blobs(t *testing.T) [][]float64 {
	t.Helper()

	//nolint:gosec // a fixed seed is what makes the fixture reproducible.
	r := rand.New(rand.NewSource(1))

	var data [][]float64

	for _, c := range [][]float64{{0, 0}, {10, 0}, {0, 10}} {
		data = append(data, gaussianBlob(r, c, 0.3, 30)...)
	}

	return data
}

// variableDensity returns two dense groups close together and one sparse group apart, which is the
// set no single link distance can cluster correctly.
//
// The dense pair sits 0.35 apart while the sparse group needs about 0.4 to connect at all, so any
// distance that holds the sparse group together also chains the dense pair into one. Labels are
// returned alongside, in the same order as the points.
func variableDensity(t *testing.T) ([][]float64, []int) {
	t.Helper()

	//nolint:gosec // a fixed seed is what makes the fixture reproducible.
	r := rand.New(rand.NewSource(7))

	var (
		data  [][]float64
		truth []int
	)

	for i, g := range []struct {
		center []float64
		sd     float64
		n      int
	}{
		{center: []float64{0, 0}, sd: 0.02, n: 40},
		{center: []float64{0.35, 0}, sd: 0.02, n: 40},
		{center: []float64{20, 20}, sd: 1.0, n: 40},
	} {
		data = append(data, gaussianBlob(r, g.center, g.sd, g.n)...)

		for range g.n {
			truth = append(truth, i+1)
		}
	}

	return data, truth
}

// separatesGroups reports whether an assignment gives each ground-truth group a cluster of its own
// holding most of it, and mixes none of them.
//
// The predicate the fixtures are read through, rather than a cluster count: a run that merges two
// groups and splits a third reports the right number of clusters and none of the right ones.
func separatesGroups(labels Labels, truth []int, groups, least int) bool {
	if labels.Count() != groups || impurity(labels, truth) != 0 {
		return false
	}

	for g := 1; g <= groups; g++ {
		if _, size := clusterOf(labels, truth, g); size < least {
			return false
		}
	}

	return true
}

// clusterOf returns the label holding the majority of the given ground-truth group, and how many
// of its points that label holds.
func clusterOf(labels Labels, truth []int, group int) (int, int) {
	counts := make(map[int]int)

	for i, g := range truth {
		if g == group && labels[i] > 0 {
			counts[labels[i]]++
		}
	}

	best, size := Noise, 0

	for label, n := range counts {
		if n > size || n == size && label < best {
			best, size = label, n
		}
	}

	return best, size
}

// impurity returns the size of the largest group of one person's points that shares a cluster with
// a majority of somebody else's, which is the cost a merge imposes.
func impurity(labels Labels, truth []int) int {
	members := make(map[int]map[int]int)

	for i, l := range labels {
		if l < 1 {
			continue
		}

		if members[l] == nil {
			members[l] = make(map[int]int)
		}

		members[l][truth[i]]++
	}

	worst := 0

	for _, byTruth := range members {
		total, top := 0, 0

		for _, n := range byTruth {
			total += n

			if n > top {
				top = n
			}
		}

		if minority := total - top; minority > worst {
			worst = minority
		}
	}

	return worst
}

// dbscanLabels runs the package's DBSCAN and returns its assignment in the shared Labels form.
func dbscanLabels(t *testing.T, data [][]float64, minPts int, eps float64) Labels {
	t.Helper()

	c, err := DBSCAN(minPts, eps, 1, EuclideanDist)

	if err != nil {
		t.Fatal(err)
	}

	if err = c.Learn(data); err != nil {
		t.Fatal(err)
	}

	labels := make(Labels, len(data))

	for i, g := range c.Guesses() {
		if g < 1 {
			labels[i] = Noise
		} else {
			labels[i] = g
		}
	}

	return labels
}

// sameClustering reports whether two assignments group the points identically, ignoring the
// numbers the clusters happen to carry.
func sameClustering(a, b Labels) bool {
	if len(a) != len(b) {
		return false
	}

	forward, backward := make(map[int]int), make(map[int]int)

	for i := range a {
		if a[i] < 1 || b[i] < 1 {
			if a[i] < 1 != (b[i] < 1) {
				return false
			}

			continue
		}

		if v, ok := forward[a[i]]; ok && v != b[i] {
			return false
		} else if v, ok = backward[b[i]]; ok && v != a[i] {
			return false
		}

		forward[a[i]], backward[b[i]] = b[i], a[i]
	}

	return true
}

// finite reports whether every value is a real number, which the scores derived from the density
// levels have to be even where a distance was zero.
func finite(values []float64) bool {
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}

	return true
}
