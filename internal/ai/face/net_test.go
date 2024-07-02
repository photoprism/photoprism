package face

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fastwalk"
)

var modelPath, _ = filepath.Abs("../../../assets/facenet")

func TestNet(t *testing.T) {
	expected := map[string]int{
		"1.jpg":  1,
		"2.jpg":  1,
		"3.jpg":  1,
		"4.jpg":  1,
		"5.jpg":  1,
		"6.jpg":  1,
		"7.jpg":  0,
		"8.jpg":  0,
		"9.jpg":  0,
		"10.jpg": 0,
		"11.jpg": 0,
		"12.jpg": 1,
		"13.jpg": 0,
		"14.jpg": 0,
		"15.jpg": 0,
		"16.jpg": 1,
		"17.jpg": 1,
		"18.jpg": 2,
		"19.jpg": 0,
	}

	faceIndices := map[string][]int{
		"18.jpg": {1, 0},
		"1.jpg":  {2},
		"4.jpg":  {3},
		"5.jpg":  {4},
		"6.jpg":  {5},
		"2.jpg":  {6},
		"12.jpg": {7},
		"16.jpg": {8},
		"17.jpg": {9},
		"3.jpg":  {10},
	}

	faceIndexToPersonID := [11]int{
		0, 1, 1, 1, 2, 0, 1, 0, 0, 1, 0,
	}

	var embeddings = make(Embeddings, 11)

	faceNet := NewNet(modelPath, "testdata/cache", false)

	if err := fastwalk.Walk("testdata", func(fileName string, info os.FileMode) error {
		if info.IsDir() || filepath.Base(filepath.Dir(fileName)) != "testdata" {
			return nil
		}

		t.Run(fileName, func(t *testing.T) {
			baseName := filepath.Base(fileName)

			faces, err := faceNet.Detect(fileName, 20, false, -1)

			if err != nil {
				t.Fatal(err)
			}

			// for i, f := range faces {
			// 	t.Logf("FACE %d IN %s: %#v", i, fileName, f.Area)
			// }

			if len(faces) > 0 {
				for i, f := range faces {
					if len(f.Embeddings) > 0 {
						// t.Logf("FACE %d IN %s: %#v", i, fileName, f.Embeddings)
						embeddings[faceIndices[baseName][i]] = f.Embeddings[0]
					} else {
						embeddings[faceIndices[baseName][i]] = nil
					}
				}
			}

			if i, ok := expected[baseName]; ok {
				assert.Equal(t, i, faces.Count())

				if faces.Count() == 0 {
					assert.Equal(t, 100, faces.Uncertainty())
				} else {
					assert.Truef(t, faces.Uncertainty() >= 0 && faces.Uncertainty() <= 50, "uncertainty should be between 0 and 50")
				}
				t.Logf("uncertainty: %d", faces.Uncertainty())
			} else {
				t.Logf("unknown test result for %s", baseName)
			}
		})

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	// Dist Matrix
	correct := 0

	for i := 0; i < len(embeddings); i++ {
		for j := 0; j < len(embeddings); j++ {
			if i >= j {
				continue
			}

			dist := embeddings[i].Dist(embeddings[j])

			t.Logf("Dist for %d %d (faces are %d %d) is %f", i, j, faceIndexToPersonID[i], faceIndexToPersonID[j], dist)
			if faceIndexToPersonID[i] == faceIndexToPersonID[j] {
				if dist < 1.21 {
					correct += 1
				}
			} else {
				if dist >= 1.21 {
					correct += 1
				}
			}
		}
	}

	t.Logf("Correct for %d", correct)

	// there are a few incorrect results
	// 3 out of 55 with the 1.21 threshold
	assert.Equal(t, 52, correct)
}
