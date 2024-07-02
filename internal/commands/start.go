package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/sevlyar/go-daemon"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/auth/session"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism/backup"
	"github.com/photoprism/photoprism/internal/server"
	"github.com/photoprism/photoprism/internal/workers"
	"github.com/photoprism/photoprism/internal/workers/auto"
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

// startAction starts the Web server and initializes the daemon.
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
			{"http-cache-public", fmt.Sprintf("%t", conf.HttpCachePublic())},
			{"http-cache-maxage", fmt.Sprintf("%d", conf.HttpCacheMaxAge())},
			{"http-video-maxage", fmt.Sprintf("%d", conf.HttpVideoMaxAge())},
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

		child, contextErr := dctx.Reborn()

		if contextErr != nil {
			return fmt.Errorf("daemon context error: %w", contextErr)
		}

		if child != nil {
			if writeErr := fs.WriteString(conf.PIDFilename(), strconv.Itoa(child.Pid)); writeErr != nil {
				log.Error(writeErr)
				return fmt.Errorf("failed writing process id to %s", clean.Log(conf.PIDFilename()))
			}

			log.Infof("daemon started with process id %v\n", child.Pid)
			return nil
		}
	}

	// Show info if read-only mode is enabled.
	if conf.ReadOnly() {
		log.Infof("config: enabled read-only mode")
	}

	// Start built-in web server.
	go server.Start(cctx, conf)

	// Restore albums from YAML files.
	if count, restoreErr := backup.RestoreAlbums(conf.BackupAlbumsPath(), false); restoreErr != nil {
		log.Errorf("restore: %s (albums)", restoreErr)
	} else if count > 0 {
		log.Infof("restore: %s restored", english.Plural(count, "album backup", "album backups"))
	}

	// Start worker that periodically deletes expired sessions.
	session.Cleanup(conf.SessionCacheDuration() * 4)

	// Start sync and metadata maintenance background workers.
	workers.Start(conf)

	// Start auto-indexing background worker.
	auto.Start(conf)

	// Wait for signal to initiate server shutdown.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	sig := <-quit

	// Stop all background activity.
	auto.Shutdown()
	workers.Shutdown()
	session.Shutdown()
	mutex.CancelAll()

	log.Info("shutting down...")
	cancel()

	if contextErr := dctx.Release(); contextErr != nil {
		log.Error(contextErr)
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
