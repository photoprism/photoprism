package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/server"
	"github.com/photoprism/photoprism/internal/util"
	"github.com/sevlyar/go-daemon"
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
	cli.BoolFlag{
		Name:   "detach-server, d",
		Usage:  "detach from the console (daemon mode)",
		EnvVar: "PHOTOPRISM_DETACH_SERVER",
	},
	cli.BoolFlag{
		Name:   "config, c",
		Usage:  "show config",
	},
}

func startAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)

	if err := conf.CreateDirectories(); err != nil {
		return err
	}

	if ctx.IsSet("config") {
		fmt.Printf("NAME                  VALUE\n")
		fmt.Printf("detach-server         %t\n", conf.DetachServer())

		fmt.Printf("sql-host              %s\n", conf.SqlServerHost())
		fmt.Printf("sql-port              %d\n", conf.SqlServerPort())
		fmt.Printf("sql-password          %s\n", conf.SqlServerPassword())
		fmt.Printf("sql-path              %s\n", conf.SqlServerPath())

		fmt.Printf("http-host             %s\n", conf.HttpServerHost())
		fmt.Printf("http-port             %d\n", conf.HttpServerPort())
		fmt.Printf("http-mode             %s\n", conf.HttpServerMode())

		return nil
	}

	// pass this context down the chain
	cctx, cancel := context.WithCancel(context.Background())

	if conf.HttpServerPort() < 1 || conf.HttpServerPort() > 65535 {
		log.Fatal("server port must be a number between 1 and 65535")
	}

	if err := conf.CreateDirectories(); err != nil {
		log.Fatal(err)
	}

	if err := conf.Init(cctx); err != nil {
		log.Fatal(err)
	}

	conf.MigrateDb()

	dctx := new(daemon.Context)
	dctx.LogFileName = conf.LogFilename()
	dctx.PidFileName = conf.PIDFilename()
	dctx.Args = ctx.Args()

	if !daemon.WasReborn() && conf.DetachServer() {
		conf.Shutdown()
		cancel()

		if pid, ok := childAlreadyRunning(conf.PIDFilename()); ok {
			log.Infof("daemon already running with process id %v\n", pid)
			return nil
		}

		child, err := dctx.Reborn()
		if err != nil {
			log.Fatal(err)
		}

		if child != nil {
			if !util.Overwrite(conf.PIDFilename(), []byte(strconv.Itoa(child.Pid))) {
				log.Fatalf("failed writing process id to \"%s\"", conf.PIDFilename())
			}

			log.Infof("daemon started with process id %v\n", child.Pid)
			return nil
		}
	}

	log.Infof("starting web server at %s:%d", conf.HttpServerHost(), conf.HttpServerPort())

	if conf.ReadOnly() {
		log.Infof("read-only mode enabled")
	}

	go server.Start(cctx, conf)

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info("shutting down...")
	conf.Shutdown()
	cancel()
	err := dctx.Release()

	if err != nil {
		log.Error(err)
	}

	time.Sleep(3 * time.Second)

	return nil
}
