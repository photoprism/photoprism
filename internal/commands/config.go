package commands

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/hub"
)

// InitConfig initializes the command config.
var InitConfig = func(ctx *cli.Context) (*config.Config, error) {
	c := config.NewConfig(ctx)
	get.SetConfig(c)
	return c, c.Init()
}

// InitCoreConfig initializes the command core config without connecting to the
// database. When quiet is true, the log level is raised to fatal so only the
// command's own output is printed (used by the config report commands).
var InitCoreConfig = func(ctx *cli.Context, quiet bool) (*config.Config, error) {
	c := config.NewConfig(ctx)
	if quiet {
		c.SetLogLevel(logrus.FatalLevel)
	}
	get.SetConfig(c)
	hub.Disable()
	return c, c.InitCore()
}
