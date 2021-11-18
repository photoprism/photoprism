package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestPlaces_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	w := NewPlaces(config.TestConfig())

	updated, err := w.Start()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("updated: %#v", updated)
}

func TestPlaces_UpdatePhotos(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("success", func(t *testing.T) {
		w := NewPlaces(config.TestConfig())
		affected, err := w.UpdatePhotos()

		if err != nil {
			t.Fatal(err)
		}

		if affected < 0 {
			t.Fatal("affected must not be negative")
		}
	})
}
