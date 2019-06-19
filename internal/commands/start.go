package commands

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/server"
	"github.com/photoprism/photoprism/internal/util"
	daemon "github.com/sevlyar/go-daemon"
	log "github.com/sirupsen/logrus"
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

	cli.BoolFlag{
		Name:   "daemonize, d",
		Usage:  "run Photoprism as Daemon",
		EnvVar: "PHOTOPRISM_DAEMON_MODE",
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

	if !daemon.WasReborn() && conf.ShouldDaemonize() {
		dctx := new(daemon.Context)
		// TODO(ved): add log file
		dctx.LogFileName = "/srv/log.txt"
		dctx.Args = ctx.Args()
		conf.Shutdown()
		cancel()
		child, err := dctx.Reborn()
		if err != nil {
			log.Fatal(err)
		}

		if child != nil {
			if !util.Overwrite(conf.DaemonPath(), []byte(strconv.Itoa(child.Pid))) {
				log.Fatal("failed to write PID to file")
			}

			log.Infof("Daemon started with PID: %v\n", child.Pid)
			return nil
		}

		defer dctx.Release()
	}

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
