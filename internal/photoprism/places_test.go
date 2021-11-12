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
