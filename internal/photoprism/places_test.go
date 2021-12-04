package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestPlaces(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	w := NewPlaces(config.TestConfig())

	t.Run("Start", func(t *testing.T) {
		updated, err := w.Start()

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("updated: %#v", updated)

		affected, err := w.UpdatePhotos()

		if err != nil {
			t.Fatal(err)
		}

		if affected < 0 {
			t.Fatal("affected must not be negative")
		}
	})
}
