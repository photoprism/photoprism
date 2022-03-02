package photoprism

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	if err := os.Remove(".test.db"); err == nil {
		log.Debugln("removed .test.db")
	}

	c := config.TestConfig()
	SetConfig(c)

	code := m.Run()

	_ = c.CloseDb()

	os.Exit(code)
}
