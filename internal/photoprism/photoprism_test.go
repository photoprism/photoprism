package photoprism

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	c := config.NewTestConfig("photoprism")
	SetConfig(c)
	defer c.CloseDb()

	code := m.Run()

	os.Exit(code)
}
