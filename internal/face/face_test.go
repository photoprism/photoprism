package face

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"math"

	"github.com/photoprism/photoprism/pkg/fastwalk"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

var assetsPath = fs.Abs("../../assets")
var modelPath = assetsPath + "/facenet"

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

	// 0 ,1 ,2,3,4,5,6,7 ,8 ,9 ,10
	// 18,18,1,4,5,6,2,12,16,17,3

	// 0,10 pair
	face_index_to_person_id := [11]int{0, 1, 1, 1, 2, 0, 1, 0, 0, 1, 0}

	var embeddings [][]float32
	tfInstance := New(assetsPath, false)
	tfInstance.loadModel()

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

			t.Logf("Found %d faces in '%s'", len(faces), baseName)

			if len(faces) > 0 {
				t.Logf("results: %#v", faces)

				for i, f := range faces {
					t.Logf("marker[%d]: %#v %#v", i, f.Marker(), f.Face)
					t.Logf("landmarks[%d]: %s", i, f.RelativeLandmarksJSON())

					embedding := tfInstance.getFaceEmbedding(fileName, f.Face)
					embeddings = append(embeddings, embedding[0])
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
	for i:=0; i<len(embeddings); i++ {
		for j:=0; j<len(embeddings); j++ {
			if i >= j {
				continue
			}
			var dist float64
			// TODO use more efficient implementation
			// either with TF or some go library, and batch processing
			for k:=0; k<512; k++ {
				dist += math.Pow(float64(embeddings[i][k] - embeddings[j][k]), 2)
			}

			math.Sqrt(dist)
			t.Logf("Dist for %d %d is %f", i, j, dist)

			if face_index_to_person_id[i] == face_index_to_person_id[j] {
				if dist < 1.5 {
					correct += 1
				}
			} else {
				if dist >= 1.5 {
					correct += 1
				}
			}
		}
	}

	// there are a few incorrect results
	// 4 out of 55 with the 1.5 threshold
	assert.True(t, correct == 51)

}
