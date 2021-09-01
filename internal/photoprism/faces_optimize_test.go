package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestFaces_Optimize(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	r, err := m.Optimize()

	if err != nil {
		t.Fatal(err)
	}

	t.Log(r)
}
