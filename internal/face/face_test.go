package face

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/photoprism/photoprism/pkg/fastwalk"
	"github.com/stretchr/testify/assert"
)

var modelPath, _ = filepath.Abs("../../assets/facenet")

func TestDetect(t *testing.T) {
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

	faceindices := map[string][]int{
		"18.jpg": {0, 1},
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

	faceindexToPersonid := [11]int{
		0, 1, 1, 1, 2, 0, 1, 0, 0, 1, 0,
	}

	var embeddings [11][]float32

	tfInstance := NewNet(modelPath, false)

	if err := tfInstance.loadModel(); err != nil {
		t.Fatal(err)
	}

	if err := fastwalk.Walk("testdata", func(fileName string, info os.FileMode) error {
		if info.IsDir() || strings.HasPrefix(filepath.Base(fileName), ".") {
			return nil
		}

		t.Run(fileName, func(t *testing.T) {
			baseName := filepath.Base(fileName)

			faces, err := Detect(fileName)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("found %d faces in '%s'", len(faces), baseName)

			if len(faces) > 0 {
				t.Logf("results: %#v", faces)

				for i, f := range faces {
					t.Logf("marker[%d]: %#v %#v", i, f.Marker(), f.Face)
					t.Logf("landmarks[%d]: %s", i, f.RelativeLandmarksJSON())

					embedding := tfInstance.getFaceEmbedding(fileName, f.Face)

					if b, err := json.Marshal(embedding[0]); err != nil {
						t.Fatal(err)
					} else {
						t.Logf("embedding: %#v", string(b))
					}

					t.Logf("face: %d %v", i, faceindices[baseName])

					embeddings[faceindices[baseName][i]] = embedding[0]
					// t.Logf("face: created embedding of face %v", embeddings[len(embeddings)-1])
				}
			}

			if i, ok := expected[baseName]; ok {
				assert.Equal(t, i, len(faces))
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

	// Distance Matrix
	correct := 0

	for i := 0; i < len(embeddings); i++ {
		for j := 0; j < len(embeddings); j++ {
			if i >= j {
				continue
			}
			dist := EuclidianDistance(embeddings[i], embeddings[j])
			t.Logf("Dist for %d %d (faces are %d %d) is %f", i, j, faceindexToPersonid[i], faceindexToPersonid[j], dist)
			if faceindexToPersonid[i] == faceindexToPersonid[j] {
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
	// 4 out of 55 with the 1.21 threshold
	assert.True(t, correct == 51)
}
