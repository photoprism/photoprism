package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestFaces_Optimize(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		c := config.TestConfig()

		m := NewFaces(c)

		r, err := m.Optimize(true)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(r)
	})

	t.Run("false", func(t *testing.T) {
		c := config.TestConfig()

		m := NewFaces(c)

		r, err := m.Optimize(false)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(r)
	})
}
