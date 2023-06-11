package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/auto"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/server"
	"github.com/photoprism/photoprism/internal/session"
	"github.com/photoprism/photoprism/internal/workers"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/report"
)

// StartCommand configures the command name, flags, and action.
var StartCommand = cli.Command{
	Name:    "start",
	Aliases: []string{"up"},
	Usage:   "Starts the Web server",
	Flags:   startFlags,
	Action:  startAction,
}

// startFlags specifies the start command parameters.
var startFlags = []cli.Flag{
	cli.BoolFlag{
		Name:   "detach-server, d",
		Usage:  "detach from the console (daemon mode)",
		EnvVar: "PHOTOPRISM_DETACH_SERVER",
	},
	cli.BoolFlag{
		Name:  "config, c",
		Usage: "show config",
	},
}

// startAction starts the web server and initializes the daemon.
func startAction(ctx *cli.Context) error {
	conf, err := InitConfig(ctx)

	if err != nil {
		return err
	}

	if ctx.IsSet("config") {
		// Create config report.
		cols := []string{"Name", "Value"}
		rows := [][]string{
			{"detach-server", fmt.Sprintf("%t", conf.DetachServer())},
			{"http-mode", conf.HttpMode()},
			{"http-compression", conf.HttpCompression()},
			{"http-cache-maxage", fmt.Sprintf("%d", conf.HttpCacheMaxAge())},
			{"http-cache-public", fmt.Sprintf("%t", conf.HttpCachePublic())},
			{"http-host", conf.HttpHost()},
			{"http-port", fmt.Sprintf("%d", conf.HttpPort())},
		}

		// Render config report.
		opt := report.Options{Format: report.CliFormat(ctx), NoWrap: true}
		result, _ := report.Render(rows, cols, opt)
		fmt.Printf("\n%s\n", result)
		return nil
	}

	if conf.HttpPort() < 1 || conf.HttpPort() > 65535 {
		log.Fatal("server port must be a number between 1 and 65535")
	}

	// Pass this context down the chain.
	cctx, cancel := context.WithCancel(context.Background())

	// Initialize the index database.
	conf.InitDb()

	// Check if the daemon is running, if not, initialize the daemon.
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
			if !fs.Overwrite(conf.PIDFilename(), []byte(strconv.Itoa(child.Pid))) {
				log.Fatalf("failed writing process id to %s", clean.Log(conf.PIDFilename()))
			}

			log.Infof("daemon started with process id %v\n", child.Pid)
			return nil
		}
	}

	if conf.ReadOnly() {
		log.Infof("config: enabled read-only mode")
	}

	// Start web server.
	go server.Start(cctx, conf)

	if count, err := photoprism.RestoreAlbums(conf.AlbumsPath(), false); err != nil {
		log.Errorf("restore: %s", err)
	} else if count > 0 {
		log.Infof("%d albums restored", count)
	}

	// Start background workers.
	session.Monitor(time.Hour)
	workers.Start(conf)
	auto.Start(conf)

	// Wait for signal to initiate server shutdown.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	sig := <-quit

	// Stop all background activity.
	auto.Stop()
	workers.Stop()
	session.Shutdown()
	mutex.CancelAll()

	log.Info("shutting down...")
	cancel()

	if err := dctx.Release(); err != nil {
		log.Error(err)
	}

	// Finally, close the DB connection after a short grace period.
	time.Sleep(2 * time.Second)
	conf.Shutdown()

	// Don't exit with 0 if SIGUSR1 was received to avoid restarts.
	if sig == syscall.SIGUSR1 {
		os.Exit(1)
		return nil
	}

	return nil
}
