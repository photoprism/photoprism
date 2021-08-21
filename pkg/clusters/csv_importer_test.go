package clusters

import (
	"math"
	"testing"
)

const TOLERANCE = 0.000001

func TestImportedLoadDataOfCorrectLengh(t *testing.T) {
	var (
		f = "data/test.csv"
		i = CsvImporter()
		s = 3
	)

	d, e := i.Import(f, 0, 2)
	if e != nil {
		t.Errorf("Error importing data: %s\n", e.Error())
	}

	if s != len(d) {
		t.Errorf("Imported data size mismatch: %d vs %d\n", s, len(d))
	}
}

func TestImportedLoadCorrectData(t *testing.T) {
	var (
		f = "data/test.csv"
		i = CsvImporter()
		s = [][]float64{
			{0.1, 0.2, 0.3},
			{0.4, 0.5, 0.6},
			{0.7, 0.8, 0.9},
		}
	)

	d, e := i.Import(f, 0, 2)
	if e != nil {
		t.Errorf("Error importing data: %s\n", e.Error())
	}

	if !fsliceEqual(d, s) {
		t.Errorf("Imported data mismatch: %v vs %v\n", d, s)
	}
}

func fsliceEqual(a, b [][]float64) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if len(a[i]) != len(b[i]) {
			return false
		}

		for j := 0; j < len(a[i]); j++ {
			if d := math.Abs(a[i][j] - b[i][j]); d > TOLERANCE {
				return false
			}
		}
	}

	return true
}

func BenchmarkImport(b *testing.B) {
	var (
		f = "data/bus-stops.csv"
		i = CsvImporter()
	)

	b.ResetTimer()

	_, e := i.Import(f, 4, 5)
	if e != nil {
		b.Errorf("Error importing data: %s\n", e.Error())
	}
}
