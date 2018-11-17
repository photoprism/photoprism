package commands

import (
	"fmt"
	"log"

	"github.com/photoprism/photoprism/internal/context"
	"github.com/photoprism/photoprism/internal/server"
	"github.com/urfave/cli"
)

var StartCommand = cli.Command{
	Name:   "start",
	Usage:  "Starts web server",
	Flags:  startFlags,
	Action: startAction,
}

var startFlags = []cli.Flag{
	cli.IntFlag{
		Name:   "server-port, p",
		Usage:  "HTTP server port",
		Value:  80,
		EnvVar: "PHOTOPRISM_SERVER_PORT",
	},
	cli.StringFlag{
		Name:   "server-host, i",
		Usage:  "HTTP server host",
		Value:  "",
		EnvVar: "PHOTOPRISM_SERVER_HOST",
	},
	cli.StringFlag{
		Name:   "server-mode, m",
		Usage:  "debug, release or test",
		Value:  "",
		EnvVar: "PHOTOPRISM_SERVER_MODE",
	},
}

// Starts web serve using startFlags; called by startCommand
func startAction(ctx *cli.Context) error {
	conf := context.NewConfig(ctx)

	if conf.GetServerPort() < 1 {
		log.Fatal("Server port must be a positive integer")
	}

	if err := conf.CreateDirectories(); err != nil {
		log.Fatal(err)
	}

	conf.MigrateDb()

	fmt.Printf("Starting web server at %s:%d...\n", ctx.String("server-host"), ctx.Int("server-port"))

	server.Start(conf)

	fmt.Println("Done.")

	return nil
}
