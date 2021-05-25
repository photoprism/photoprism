package face

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/photoprism/photoprism/pkg/fastwalk"
	"github.com/stretchr/testify/assert"
)

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
					t.Logf("marker[%d]: %#v", i, f.Marker())
					t.Logf("landmarks[%d]: %s", i, f.RelativeLandmarksJSON())
				}
			}

			if i, ok := expected[baseName]; ok {
				assert.Equal(t, i, len(faces))
			} else {
				t.Errorf("unknown test result for %s", baseName)
			}
		})

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
