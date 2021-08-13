package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestPeople_Start(t *testing.T) {
	conf := config.TestConfig()

	m := NewFaces(conf)
	err := m.Start()

	if err != nil {
		t.Fatal(err)
	}
}
