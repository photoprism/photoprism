package photoprism

import (
	"github.com/photoprism/photoprism/internal/config"
	"testing"
)

func TestFaces_Reset(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	err := m.Reset()

	if err != nil {
		t.Fatal(err)
	}
}
