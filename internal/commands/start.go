package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/sevlyar/go-daemon"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/session"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism/backup"
	"github.com/photoprism/photoprism/internal/server"
	"github.com/photoprism/photoprism/internal/server/process"
	"github.com/photoprism/photoprism/internal/workers"
	"github.com/photoprism/photoprism/internal/workers/auto"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// StartCommand configures the command name, flags, and action.
var StartCommand = &cli.Command{
	Name:    "start",
	Aliases: []string{"up"},
	Usage:   "Starts the Web server",
	Flags:   startFlags,
	Action:  startAction,
}

// startFlags specifies the start command parameters.
var startFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "detach-server",
		Aliases: []string{"d"},
		Usage:   "detach from the console (daemon mode)",
		EnvVars: config.EnvVars("DETACH_SERVER"),
	},
	&cli.BoolFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "show config",
	},
}

// daemonPidTimeout limits how long the start command waits for the daemon to write its process ID.
const daemonPidTimeout = time.Minute

// writePidFd writes the process ID to the pid file the daemon inherited.
var writePidFd = syscall.Write

// daemonContextMaxSize limits how much of the context go-daemon passes on its standard input is read.
const daemonContextMaxSize = 4 << 20

// daemonContext holds the part of the context that go-daemon passes to the daemon on its standard input.
type daemonContext struct {
	PidFileName string
}

// startPrologue prepares the process before the start command loads its config. A daemon that cannot write its
// process ID to the pid file returns an error, so that it exits.
func startPrologue() error {
	if daemon.WasReborn() {
		// Programs the daemon runs do not inherit its pid file.
		closePidFileOnExec()

		pidFile, err := daemonPidFileName(os.Stdin)

		if err == nil {
			err = writeDaemonPid(pidFile)
		}

		if err != nil {
			return fmt.Errorf("daemon: failed to write process id to pid file (%w)", err)
		}
	}

	// Leave audit persistence to the event hub, so that requests do not wait for audit writes.
	event.SetAuditRecorder(nil)

	return nil
}

// daemonPidFileName returns the pid file name from the context that go-daemon passes to the daemon through its
// standard input, which must be a pipe.
func daemonPidFileName(stdin *os.File) (string, error) {
	var st syscall.Stat_t

	if err := syscall.Fstat(int(stdin.Fd()), &st); err != nil { //nolint:gosec // File descriptors fit into an int.
		return "", err
	} else if st.Mode&syscall.S_IFMT != syscall.S_IFIFO {
		return "", errors.New("standard input is not a pipe")
	}

	var dctx daemonContext

	if err := json.NewDecoder(io.LimitReader(stdin, daemonContextMaxSize)).Decode(&dctx); err != nil {
		return "", fmt.Errorf("invalid daemon context (%w)", err)
	} else if dctx.PidFileName == "" {
		return "", errors.New("daemon context has no pid file name")
	}

	return dctx.PidFileName, nil
}

// clearDaemonMark removes the variable that marks the daemon, so that the programs the server runs do not inherit it.
func clearDaemonMark() {
	_ = os.Unsetenv(daemon.MARK_NAME)
}

// writeDaemonPid writes the process ID to the pid file the daemon inherited and syncs it. The inherited file must
// be the regular file at pidFile, which must not be a symbolic link, and have a single name.
func writeDaemonPid(pidFile string) error {
	var st syscall.Stat_t

	if err := syscall.Fstat(daemonPidFileFd, &st); err != nil {
		return err
	} else if st.Mode&syscall.S_IFMT != syscall.S_IFREG {
		return errors.New("pid file is not a regular file")
	} else if st.Nlink > 1 {
		return errPidFileLinked
	}

	if link, err := os.Lstat(pidFile); err != nil { //nolint:gosec // Path from the daemon context.
		return err
	} else if link.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("pid file %s must not be a symbolic link", clean.Log(pidFile))
	} else if current, ok := link.Sys().(*syscall.Stat_t); !ok || current.Dev != st.Dev || current.Ino != st.Ino {
		return fmt.Errorf("pid file %s is not the file the daemon inherited", clean.Log(pidFile))
	}

	pid := []byte(strconv.Itoa(os.Getpid()))

	// Empties the file first, so that no old content remains even if the daemon is killed while writing.
	if err := syscall.Ftruncate(daemonPidFileFd, 0); err != nil {
		return err
	}

	if _, err := syscall.Seek(daemonPidFileFd, 0, io.SeekStart); err != nil {
		return err
	}

	if n, err := writePidFd(daemonPidFileFd, pid); err != nil || n < len(pid) {
		// Removes what was written, as a part of the process ID would name another process.
		_ = syscall.Ftruncate(daemonPidFileFd, 0)

		if err == nil {
			err = io.ErrShortWrite
		}

		return err
	}

	return syscall.Fsync(daemonPidFileFd)
}

// waitForDaemonPid waits until the daemon has written its process ID to the pid file, and fails if it exits first
// or the timeout passes.
func waitForDaemonPid(pidFile string, child *os.Process, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if pid, _, err := readPidFile(pidFile); err == nil && pid == child.Pid {
			return nil
		} else if !processRunning(child) {
			return fmt.Errorf("daemon %d exited before writing its process id to %s", child.Pid, clean.Log(pidFile))
		} else if time.Now().After(deadline) {
			return fmt.Errorf("daemon %d has not written its process id to %s within %s and may still be starting", child.Pid, clean.Log(pidFile), timeout)
		}

		time.Sleep(20 * time.Millisecond)
	}
}

// daemonRunning reports whether a daemon holds the pid file, and refuses a pid file that is a symbolic link.
func daemonRunning(pidFile string) (pid int, running bool, err error) {
	if fs.IsSymlink(pidFile) {
		return 0, false, fmt.Errorf("pid file %s must not be a symbolic link", clean.Log(pidFile))
	}

	return childAlreadyRunning(pidFile)
}

// startAction starts the Web server and initializes the daemon.
func startAction(ctx *cli.Context) error {
	if err := startPrologue(); err != nil {
		return err
	}

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
			{"http-header-timeout", conf.HttpHeaderTimeout().String()},
			{"http-header-bytes", fmt.Sprintf("%d", conf.HttpHeaderBytes())},
			{"http-idle-timeout", conf.HttpIdleTimeout().String()},
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
	dctx.Args = ctx.Args().Slice()

	if !daemon.WasReborn() && conf.DetachServer() {
		conf.Shutdown()
		cancel()

		if pid, running, runningErr := daemonRunning(conf.PIDFilename()); runningErr != nil {
			return runningErr
		} else if running {
			log.Infof("daemon already running with process id %v\n", pid)
			return nil
		}

		child, contextErr := dctx.Reborn()

		if contextErr != nil {
			return fmt.Errorf("daemon context error: %w", contextErr)
		}

		if child != nil {
			// The daemon writes its process ID through the pid file it inherited.
			if waitErr := waitForDaemonPid(conf.PIDFilename(), child, daemonPidTimeout); waitErr != nil {
				return fmt.Errorf("%w, see %s", waitErr, clean.Log(conf.LogFilename()))
			}

			log.Infof("daemon started with process id %v\n", child.Pid)
			return nil
		}
	}

	clearDaemonMark()

	// Show info if read-only mode is enabled.
	if conf.ReadOnly() {
		log.Infof("config: enabled read-only mode")
	}

	warnNonASCIIUsernames()

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

	// Wait for signal to trigger server shutdown or restart.
	signal.Notify(process.Signal, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)
	sig := <-process.Signal

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

	// Exit with status code 1 if the shutdown was initiated with SIGUSR1 to request a restart.
	//
	// Note that this requires an entrypoint script or other process to
	// spawns a new instance when the server exists with status code 1.
	if sig == syscall.SIGUSR1 {
		os.Exit(1)
		return nil
	}

	return nil
}

// warnNonASCIIUsernames logs a system warning with the number of accounts whose usernames contain characters
// other than ASCII, which new and renamed accounts cannot have, and returns it.
func warnNonASCIIUsernames() int {
	count := entity.CountNonASCIIUsernames()

	if count > 0 {
		event.SystemWarn([]string{"users", "%s with characters other than ASCII in the username", "rename with photoprism users mod --username"}, english.Plural(count, "account", "accounts"))
	}

	return count
}
