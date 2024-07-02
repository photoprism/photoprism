package get

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestMain(m *testing.M) {
	c := config.NewTestConfig("service")
	SetConfig(c)
	defer c.CloseDb()

	code := m.Run()

	os.Exit(code)
}
