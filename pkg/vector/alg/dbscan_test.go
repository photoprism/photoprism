package alg

import (
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

	for _, entry := range progress {
		if entry[0] <= 0 {
			t.Fatalf("expected a positive processed count, got %d", entry[0])
		}
		if entry[1] != len(points) {
			t.Fatalf("expected total %d, got %d", len(points), entry[1])
		}
	}
}
