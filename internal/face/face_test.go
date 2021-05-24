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
		"5.jpg":  2,
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
	}

	if err := fastwalk.Walk("testdata", func(fileName string, info os.FileMode) error {
		if info.IsDir() || strings.HasPrefix(filepath.Base(fileName), ".") {
			return nil
		}

		t.Run(fileName, func(t *testing.T) {
			baseName := filepath.Base(fileName)

			res, err := Detect(fileName, DefaultDetector())

			if err != nil {
				t.Fatal(err)
			}

			t.Logf("Found %d faces in '%s'", len(res), baseName)

			if len(res) > 0 {
				t.Logf("results: %#v", res)
				for i, r := range res {
					t.Logf("landmarks[%d]: %d", i, len(r.Landmarks))
				}
				t.Logf("regions: %#v", res.Regions())
			}

			if i, ok := expected[baseName]; ok {
				assert.Equal(t, len(res), i)
			} else {
				t.Errorf("unknown test result for %s", baseName)
			}
		})

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
