package workers

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)

	if err := os.Remove(".test.db"); err == nil {
		log.Debugln("removed .test.db")
	}

	c := config.TestConfig()

	code := m.Run()

	_ = c.CloseDb()

	os.Exit(code)
}
