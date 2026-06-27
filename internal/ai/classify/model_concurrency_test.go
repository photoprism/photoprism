package classify

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestModel_RunConcurrent verifies that a single model instance can classify
// images from multiple goroutines without corrupting results, as happens during
// parallel indexing. Run with -race to deterministically catch the data race.
func TestModel_RunConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	model := NewNasnet(modelsPath, false)
	if err := model.loadModel(); err != nil {
		t.Fatal(err)
	}

	// Distinct subjects with stable top labels; if the shared input buffer is
	// clobbered by a concurrent worker, the produced label no longer matches.
	cases := []struct {
		file  string
		label string
	}{
		{"chameleon_lime.jpg", "chameleon"},
		{"dog_orange.jpg", "dog"},
		{"cat_224.jpeg", "cat"},
		{"zebra_green_brown.jpg", "zebra"},
	}

	images := make([][]byte, len(cases))
	for i := range cases {
		data, err := os.ReadFile(filepath.Join(samplesPath, cases[i].file)) //nolint:gosec // reading bundled test fixture
		if err != nil {
			t.Fatal(err)
		}
		images[i] = data
	}

	const goroutinesPerCase = 2
	const rounds = 3

	var wg sync.WaitGroup
	start := make(chan struct{})
	var mu sync.Mutex
	var mismatches []string

	for i := range cases {
		for g := 0; g < goroutinesPerCase; g++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				<-start
				for r := 0; r < rounds; r++ {
					result, err := model.Run(images[idx], 10)
					if err != nil {
						mu.Lock()
						mismatches = append(mismatches, cases[idx].file+": "+err.Error())
						mu.Unlock()
						return
					}
					if len(result) == 0 || result[0].Name != cases[idx].label {
						got := "<none>"
						if len(result) > 0 {
							got = result[0].Name
						}
						mu.Lock()
						mismatches = append(mismatches, cases[idx].file+": expected "+cases[idx].label+", got "+got)
						mu.Unlock()
					}
				}
			}(i)
		}
	}

	close(start)
	wg.Wait()

	assert.Empty(t, mismatches, "concurrent classification corrupted labels: %v", mismatches)
}
