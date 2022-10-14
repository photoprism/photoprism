package commands

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	c := config.NewTestConfig("commands")
	service.SetConfig(c)

	InitConfig = func(ctx *cli.Context) (*config.Config, error) {
		return c, c.Init()
	}

	code := m.Run()

	os.Exit(code)
}
