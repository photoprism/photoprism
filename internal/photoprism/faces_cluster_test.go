package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestFaces_Cluster(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	opt := FacesOptions{
		Force:     true,
		Threshold: 1,
	}

	r, err := m.Cluster(opt)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(r)
}
