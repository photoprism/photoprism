package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestFaces_Start(t *testing.T) {
	conf := config.TestConfig()

	m := NewFaces(conf)
	err := m.Start()

	if err != nil {
		t.Fatal(err)
	}
}
