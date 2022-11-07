package clusters

import (
	"testing"
)

func TestKmeansClusterNumberMatches(t *testing.T) {
	const (
		C = 8
	)

	var (
		f = "data/bus-stops.csv"
		i = CsvImporter()
	)

	d, e := i.Import(f, 4, 5)
	if e != nil {
		t.Errorf("Error importing data: %s\n", e.Error())
	}

	c, e := KMeans(1000, C, EuclideanDist)
	if e != nil {
		t.Errorf("Error initializing kmeans clusterer: %s\n", e.Error())
	}

	if e = c.Learn(d); e != nil {
		t.Errorf("Error learning data: %s\n", e.Error())
	}

	if len(c.Sizes()) != C {
		t.Errorf("Number of clusters does not match: %d vs %d\n", len(c.Sizes()), C)
	}
}
