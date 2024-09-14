package commands

import (
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// InitConfig initializes the command config.
var InitConfig = func(ctx *cli.Context) (*config.Config, error) {
	c := config.NewConfig(ctx)
	get.SetConfig(c)
	return c, c.Init()
}
