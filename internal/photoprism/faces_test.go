package photoprism

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
)

func TestFaces_Start(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	opt := FacesOptions{
		Force:     true,
		Threshold: 1,
	}

	err := m.Start(opt)

	if err != nil {
		t.Fatal(err)
	}
}

func TestFaces_StartDefault(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	err := m.StartDefault()

	if err != nil {
		t.Fatal(err)
	}
}

func TestFaces_start(t *testing.T) {
	c := config.TestConfig()
	m := NewFaces(c)

	require.NoError(t, m.start(FacesOptions{Force: true, Threshold: 1}))

	var invalid *Faces
	require.Error(t, invalid.start(FacesOptions{}))
}
