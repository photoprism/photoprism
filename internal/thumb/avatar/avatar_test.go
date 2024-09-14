package avatar

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	c := config.NewTestConfig("avatar")
	get.SetConfig(c)
	photoprism.SetConfig(c)
	defer c.CloseDb()

	code := m.Run()

	os.Exit(code)
}
