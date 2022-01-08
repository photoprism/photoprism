package clusters

import (
	"encoding/json"
	"os"
)

type jsonImporter struct {
}

func JsonImporter() Importer {
	return &jsonImporter{}
}

func (i *jsonImporter) Import(file string, start, end int) ([][]float64, error) {
	if start < 0 || end < 0 || start > end {
		return [][]float64{}, errInvalidRange
	}

	f, err := os.ReadFile(file)
	if err != nil {
		return [][]float64{}, err
	}

	var (
		d = make([][]float64, 0)
		s = end - start + 1
		g = make([]float64, 0, s)
		c int
	)

	err = json.Unmarshal(f, &d)
	if err != nil {
		return [][]float64{}, err
	}

	for i := range d {
		c = 0

		for j := start; j <= end; j++ {
			g[c] = d[i][j]
			c++
		}

		d[i] = make([]float64, 0, s)
		for j := 0; j < s; j++ {
			d[i][j] = g[j]
		}
	}

	return d, nil
}
