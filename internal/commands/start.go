package commands

import (
	log "github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/context"
	"github.com/photoprism/photoprism/internal/server"
	"github.com/urfave/cli"
)

// Starts web server (user interface)
var StartCommand = cli.Command{
	Name:   "start",
	Usage:  "Starts web server",
	Flags:  startFlags,
	Action: startAction,
}

var startFlags = []cli.Flag{
	cli.IntFlag{
		Name:   "http-port, p",
		Usage:  "HTTP server port",
		Value:  80,
		EnvVar: "PHOTOPRISM_HTTP_PORT",
	},
	cli.StringFlag{
		Name:   "http-host, i",
		Usage:  "HTTP server host",
		Value:  "",
		EnvVar: "PHOTOPRISM_HTTP_HOST",
	},
	cli.StringFlag{
		Name:   "http-mode, m",
		Usage:  "debug, release or test",
		Value:  "",
		EnvVar: "PHOTOPRISM_HTTP_MODE",
	},
}

func startAction(ctx *cli.Context) error {
	app := context.NewContext(ctx)

	if app.HttpServerPort() < 1 {
		log.Fatal("server port must be a positive integer")
	}

	if err := app.CreateDirectories(); err != nil {
		log.Fatal(err)
	}

	app.MigrateDb()

	log.Infof("starting web server at %s:%d", app.HttpServerHost(), app.HttpServerPort())

	server.Start(app)

	return nil
}
