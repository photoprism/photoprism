package photoprism

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)
	code := m.Run()
	os.Exit(code)
}
