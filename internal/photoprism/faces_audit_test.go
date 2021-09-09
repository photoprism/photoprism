package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestFaces_Audit(t *testing.T) {
	t.Run("fix == true", func(t *testing.T) {
		c := config.TestConfig()

		m := NewFaces(c)

		err := m.Audit(true)

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("fixe == false", func(t *testing.T) {
		c := config.TestConfig()

		m := NewFaces(c)

		err := m.Audit(false)

		if err != nil {
			t.Fatal(err)
		}
	})
}
