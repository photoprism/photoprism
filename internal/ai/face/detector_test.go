package face

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dustin/go-humanize/english"
	pigo "github.com/esimov/pigo/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs/fastwalk"
)

var benchmarkFacesCount int

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
		"17.jpg": 2,
		"18.jpg": 2,
		"19.jpg": 1,
	}

	if err := fastwalk.Walk("testdata", func(fileName string, info os.FileMode) error {
		if info.IsDir() || filepath.Base(filepath.Dir(fileName)) != "testdata" {
			return nil
		}

		t.Run(fileName, func(t *testing.T) {
			baseName := filepath.Base(fileName)

			faces, err := Detect(fileName, true, 20)

			if err != nil {
				t.Fatal(err)
			}

			t.Logf("found %s in '%s'", english.Plural(len(faces), "face", "faces"), baseName)

			if len(faces) > 0 {
				// t.Logf("results: %#v", faces)

				for i, f := range faces {
					t.Logf("marker[%d]: %#v %#v", i, f.CropArea(), f.Area)
					t.Logf("landmarks[%d]: %s", i, f.RelativeLandmarksJSON())
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
}

func TestDetectOverlap(t *testing.T) {
	expected := map[string]int{
		"1.jpg": 2,
		"2.jpg": 2,
		"3.jpg": 2,
		"4.jpg": 1,
	}

	if err := fastwalk.Walk("testdata/overlap", func(fileName string, info os.FileMode) error {
		if info.IsDir() || filepath.Base(filepath.Dir(fileName)) != "overlap" {
			return nil
		}

		t.Run(fileName, func(t *testing.T) {
			baseName := filepath.Base(fileName)

			faces, err := Detect(fileName, true, 20)

			if err != nil {
				t.Fatal(err)
			}

			t.Logf("found %s in '%s'", english.Plural(len(faces), "face", "faces"), baseName)

			if len(faces) > 0 {
				// t.Logf("results: %#v", faces)

				for i, f := range faces {
					t.Logf("marker[%d]: %#v %#v", i, f.CropArea(), f.Area)
					t.Logf("landmarks[%d]: %s", i, f.RelativeLandmarksJSON())
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
}

func TestDetectLandmarkCounts(t *testing.T) {
	faces, err := Detect("testdata/18.jpg", true, 20)
	require.NoError(t, err)
	require.Equal(t, 2, faces.Count())

	expectedEyes := []int{2, 0}
	expectedLandmarks := []int{15, 0}

	for i, face := range faces {
		t.Run(fmt.Sprintf("face-%d", i), func(t *testing.T) {
			t.Logf("eyes=%d landmarks=%d", len(face.Eyes), len(face.Landmarks))
			require.Equal(t, expectedEyes[i], len(face.Eyes))
			require.Equal(t, expectedLandmarks[i], len(face.Landmarks))
		})
	}
}

func TestDetectQualityFallback(t *testing.T) {
	t.SkipNow()
	faces, err := Detect("testdata/<no public test image available>.jpg", false, 20)
	require.NoError(t, err)
	require.NotEmpty(t, faces)

	found := false

	for _, face := range faces {
		if face.Score < int(QualityThreshold(face.Area.Scale)) {
			found = true
			break
		}
	}

	require.Truef(t, found, "expected at least one face below the quality threshold, got %+v", faces)
}

func BenchmarkDetectorFacesLandmarks(b *testing.B) {
	const sample = "testdata/18.jpg"

	d := &pigoDetector{
		minSize:       20,
		shiftFactor:   0.1,
		scaleFactor:   1.1,
		iouThreshold:  float64(OverlapThresholdFloor) / 100,
		perturb:       63,
		landmarkAngle: 0.0,
		angles:        append([]float64(nil), DetectionAngles...),
	}

	det, params, err := d.Detect(sample)
	if err != nil {
		b.Fatal(err)
	}

	if len(det) == 0 {
		b.Fatalf("no detections found for %s", sample)
	}

	b.ReportAllocs()
	b.ResetTimer()

	detections := make([]pigo.Detection, len(det))

	for b.Loop() {
		copy(detections, det)

		faces, err := d.Faces(detections, params, true)
		if err != nil {
			b.Fatal(err)
		}

		benchmarkFacesCount = faces.Count()
	}
}
