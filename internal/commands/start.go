package commands

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
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
	// pass this context down the chain
	cctx, cancel := context.WithCancel(context.Background())
	conf := config.NewConfig(ctx)
	if conf.HttpServerPort() < 1 {
		log.Fatal("server port must be a positive integer")
	}

	if err := conf.CreateDirectories(); err != nil {
		log.Fatal(err)
	}

	if err := conf.Init(cctx); err != nil {
		log.Fatal(err)
	}
	conf.MigrateDb()

	log.Infof("starting web server at %s:%d", conf.HttpServerHost(), conf.HttpServerPort())
	go server.Start(cctx, conf)

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info("Shutting down...")
	conf.Shutdown()
	cancel()
	time.Sleep(3 * time.Second)
	return nil
}
