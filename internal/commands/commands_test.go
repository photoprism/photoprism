package commands

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/config"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	c := config.NewTestConfig("commands")

	code := m.Run()

	_ = c.CloseDb()

	os.Exit(code)
}
