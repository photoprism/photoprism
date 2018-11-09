package commands

import (
	"fmt"
	"log"

	"github.com/photoprism/photoprism/internal/photoprism"
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
func startAction(context *cli.Context) error {
	conf := photoprism.NewConfig(context)

	if conf.GetServerPort() < 1 {
		log.Fatal("Server port must be a positive integer")
	}

	if err := conf.CreateDirectories(); err != nil {
		log.Fatal(err)
	}

	conf.MigrateDb()

	fmt.Printf("Starting web server at %s:%d...\n", context.String("server-host"), context.Int("server-port"))

	server.Start(conf)

	fmt.Println("Done.")

	return nil
}
