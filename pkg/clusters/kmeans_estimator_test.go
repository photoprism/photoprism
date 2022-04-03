package clusters

import (
	"testing"
)

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
