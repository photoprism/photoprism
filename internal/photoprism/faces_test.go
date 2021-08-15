package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestFaces_Start(t *testing.T) {
	conf := config.TestConfig()

	m := NewFaces(conf)

	opt := FacesOptions{
		Force: true,
	}

	err := m.Start(opt)

	if err != nil {
		t.Fatal(err)
	}
}
