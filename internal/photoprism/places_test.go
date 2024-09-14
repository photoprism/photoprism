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

	t.Run("Unresolved", func(t *testing.T) {
		updated, err := w.Start(false)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("updated: %#v", updated)

		affected, err := w.UpdatePhotos(false)

		if err != nil {
			t.Fatal(err)
		}

		if affected < 0 {
			t.Fatal("affected must not be negative")
		}
	})

	t.Run("Force", func(t *testing.T) {
		updated, err := w.Start(true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("updated: %#v", updated)

		affected, err := w.UpdatePhotos(true)

		if err != nil {
			t.Fatal(err)
		}

		if affected < 0 {
			t.Fatal("affected must not be negative")
		}
	})
}
