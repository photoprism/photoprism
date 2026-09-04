package alg

import (
	"testing"
)

func TestKmeansEstimatorRaggedData(t *testing.T) {
	// Estimate calls bounds, which reads every row at the width of the first one from a
	// goroutine, so ragged input is rejected before it gets there.
	c, err := KMeansEstimator(10, 2, EuclideanDist)
	if err != nil {
		t.Fatalf("unexpected constructor error: %s", err)
	}

	if _, err = c.Estimate([][]float64{{1, 1}, {2}}); err != errRaggedData {
		t.Errorf("expected errRaggedData, got %v", err)
	}
}

func TestKmeansEstimator(t *testing.T) {
	const (
		C = 10
		E = 1
	)

	var (
		f = "data/bus-stops.csv"
		i = CsvImporter()
	)

	d, e := i.Import(f, 4, 5)
	if e != nil {
		t.Errorf("Error importing data: %s\n", e.Error())
	}

	c, e := KMeansEstimator(1000, C, EuclideanDist)
	if e != nil {
		t.Errorf("Error initializing kmeans clusterer: %s\n", e.Error())
	}

	r, e := c.Estimate(d)
	if e != nil {
		t.Errorf("Error running test: %s\n", e.Error())
	}

	if r != E {
		t.Errorf("Estimated number of clusters should be %d, it s %d\n", E, r)
	}
}
