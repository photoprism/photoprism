package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestMoments_Start(t *testing.T) {
	conf := config.TestConfig()

	m := NewMoments(conf)
	err := m.Start()

	if err != nil {
		t.Fatal(err)
	}
}
