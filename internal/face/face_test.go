package face

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/photoprism/photoprism/pkg/fs"

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

	tfInstance := NewNet(modelPath, "testdata/cache", false)

	if err := tfInstance.loadModel(); err != nil {
		t.Fatal(err)
	}

	if err := fastwalk.Walk("testdata", func(fileName string, info os.FileMode) error {
		if info.IsDir() || strings.HasPrefix(filepath.Base(fileName), ".") || strings.Contains(fileName, "cache") {
			return nil
		}

		t.Run(fileName, func(t *testing.T) {
			fileHash := fs.Hash(fileName)
			baseName := filepath.Base(fileName)

			faces, err := Detect(fileName, true)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("found %d faces in '%s'", len(faces), baseName)

			if len(faces) > 0 {
				t.Logf("results: %#v", faces)

				for i, f := range faces {
					t.Logf("marker[%d]: %#v %#v", i, f.Crop(), f.Area)
					t.Logf("landmarks[%d]: %s", i, f.RelativeLandmarksJSON())

					img, err := tfInstance.getFaceCrop(fileName, fileHash, &faces[i])

					if err != nil {
						t.Fatal(err)
					}

					embedding := tfInstance.getEmbeddings(img)

					if b, err := json.Marshal(embedding[0]); err != nil {
						t.Fatal(err)
					} else {
						t.Logf("embedding: %#v", string(b))
					}

					t.Logf("faces: %d %v", i, faceindices[baseName])
					embeddings[faceindices[baseName][i]] = embedding[0]
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

func TestFaces_Uncertainty(t *testing.T) {
	t.Run("maxScore = 310", func(t *testing.T) {
		f := Faces{Face{Score: 310}, Face{Score: 210}}
		assert.Equal(t, 1, f.Uncertainty())
	})
	t.Run("maxScore = 210", func(t *testing.T) {
		f := Faces{Face{Score: 210}, Face{Score: 210}}
		assert.Equal(t, 5, f.Uncertainty())
	})
	t.Run("maxScore = 66", func(t *testing.T) {
		f := Faces{Face{Score: 66}, Face{Score: 66}}
		assert.Equal(t, 20, f.Uncertainty())
	})
	t.Run("maxScore = 10", func(t *testing.T) {
		f := Faces{Face{Score: 10}, Face{Score: 10}}
		assert.Equal(t, 50, f.Uncertainty())
	})
}

func TestFace_Size(t *testing.T) {
	t.Run("8", func(t *testing.T) {
		f := Face{
			Rows:  8,
			Cols:  1,
			Score: 200,
			Area: Area{
				Name:  "",
				Row:   0,
				Col:   0,
				Scale: 8,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}
		assert.Equal(t, 8, f.Size())
	})
}

func TestFace_Dim(t *testing.T) {
	t.Run("3", func(t *testing.T) {
		f := Face{
			Rows:  8,
			Cols:  3,
			Score: 200,
			Area: Area{
				Name:  "",
				Row:   0,
				Col:   0,
				Scale: 8,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}
		assert.Equal(t, float32(3), f.Dim())
	})
	t.Run("1", func(t *testing.T) {
		f := Face{
			Rows:  8,
			Cols:  0,
			Score: 200,
			Area: Area{
				Name:  "",
				Row:   0,
				Col:   0,
				Scale: 8,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}
		assert.Equal(t, float32(1), f.Dim())
	})
}

func TestFace_EmbeddingsJSON(t *testing.T) {
	t.Run("no result", func(t *testing.T) {
		f := Face{
			Rows:  8,
			Cols:  1,
			Score: 200,
			Area: Area{
				Name:  "",
				Row:   0,
				Col:   0,
				Scale: 8,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}
		assert.Equal(t, []byte{0x6e, 0x75, 0x6c, 0x6c}, f.EmbeddingsJSON())
	})
}

func TestFace_Crop(t *testing.T) {
	t.Run("Position", func(t *testing.T) {
		f := Face{
			Cols:  1000,
			Rows:  600,
			Score: 125,
			Area: Area{
				Name:  "face",
				Col:   400,
				Row:   250,
				Scale: 200,
			},
			Eyes:       nil,
			Landmarks:  nil,
			Embeddings: nil,
		}
		t.Logf("marker: %#v", f.Crop())
	})
}
