package alg

import (
	"testing"
)

func TestKmeansRaggedData(t *testing.T) {
	// Mismatched widths reach gonum's floats.Add, which panics rather than returning an
	// error, so they are rejected before any accumulator is sized.
	c, err := KMeans(10, 2, EuclideanDist)
	if err != nil {
		t.Fatalf("unexpected constructor error: %s", err)
	}

	if err = c.Learn([][]float64{{1, 1}, {2, 2}, {3}}); err != errRaggedData {
		t.Errorf("expected errRaggedData, got %v", err)
	}
}

func TestKmeansPredict(t *testing.T) {
	c, err := KMeans(10, 2, EuclideanDist)
	if err != nil {
		t.Fatalf("unexpected constructor error: %s", err)
	}
	t.Run("Untrained", func(t *testing.T) {
		if n := c.Predict([]float64{1, 1}); n != -1 {
			t.Errorf("expected -1, got %d", n)
		}
	})
	t.Run("WrongDimensions", func(t *testing.T) {
		if err = c.Learn([][]float64{{1, 1}, {1, 2}, {9, 9}, {9, 8}}); err != nil {
			t.Fatalf("unexpected learn error: %s", err)
		}
		if n := c.Predict([]float64{1}); n != -1 {
			t.Errorf("expected -1, got %d", n)
		}
	})
}

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
