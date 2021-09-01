package photoprism

import (
	"github.com/photoprism/photoprism/internal/config"
	"testing"
)

func TestFaces_Stats(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	err := m.Stats()

	if err != nil {
		t.Fatal(err)
	}
}
