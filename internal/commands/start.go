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

// daemonPidTimeout limits how long the start command waits for a daemon to write its process ID.
var daemonPidTimeout = time.Minute

// writePidFd writes the process ID to the pid file the daemon inherited.
var writePidFd = syscall.Write

// daemonStatus describes the outcome of starting the daemon.
type daemonStatus int

const (
	// daemonStarted means that a new daemon was started.
	daemonStarted daemonStatus = iota
	// daemonAlreadyRunning means that a daemon runs and holds the pid file.
	daemonAlreadyRunning
)

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

// detachRequested reports whether this process should start the daemon and exit. Otherwise it runs the server, and
// the daemon marker is removed, so that the programs the server runs do not inherit it.
func detachRequested(detach bool) bool {
	if detach && !daemon.WasReborn() {
		return true
	}

	clearDaemonMark()

	return false
}

// detachDaemon starts the daemon unless one runs or is starting, and waits until it has written its process ID.
// While another process holds the pid file lock, it waits for that daemon, and tries again if the lock is released.
func detachDaemon(dctx *daemon.Context) (pid int, status daemonStatus, err error) {
	deadline := time.Now().Add(daemonPidTimeout)
	starting := false

	for {
		if pid, running, runningErr := daemonRunning(dctx.PidFileName); runningErr != nil {
			return 0, daemonStarted, runningErr
		} else if running {
			return pid, daemonAlreadyRunning, nil
		}

		// The pid file as it was before, so that the process ID written by the new daemon can be told apart.
		before, _ := os.Lstat(dctx.PidFileName)

		child, rebornErr := dctx.Reborn()

		switch {
		case errors.Is(rebornErr, daemon.ErrWouldBlock):
			if !starting {
				log.Info("daemon is starting")
				starting = true
			}

			pid, status, released, waitErr := waitForStartingDaemon(dctx.PidFileName, deadline)

			if released {
				continue
			}

			return pid, status, waitErr
		case rebornErr != nil:
			return 0, daemonStarted, fmt.Errorf("daemon context error: %w", rebornErr)
		case child == nil:
			return 0, daemonStarted, errors.New("daemon context error: no daemon process")
		}

		if err = waitForDaemonPid(dctx.PidFileName, before, child, daemonPidTimeout); err != nil {
			return child.Pid, daemonStarted, fmt.Errorf("%w, see %s", err, clean.Log(dctx.LogFileName))
		}

		return child.Pid, daemonStarted, nil
	}
}

// waitForStartingDaemon waits until a daemon that another process is starting holds the pid file. released reports
// that the lock was released without a daemon. After the deadline, it fails, naming the processes that have the file
// open unless the daemon cannot be confirmed.
func waitForStartingDaemon(pidFile string, deadline time.Time) (pid int, status daemonStatus, released bool, err error) {
	for {
		// Errors are expected while the daemon has not written its process ID yet.
		pid, running, runningErr := daemonRunning(pidFile)

		if runningErr == nil && running {
			return pid, daemonAlreadyRunning, false, nil
		} else if !time.Now().After(deadline) {
			time.Sleep(100 * time.Millisecond)

			// Lets the caller try again once the lock is released, but not before the pause above.
			if held, known := pidFileLocked(pidFile, nil); runningErr == nil && known && !held {
				return 0, daemonStarted, true, nil
			}

			continue
		} else if errors.Is(runningErr, errDaemonUnconfirmed) {
			return 0, daemonStarted, false, runningErr
		}

		var hint string

		if _, info, readErr := readPidFile(pidFile); info != nil || readErr == nil {
			hint = pidFileHolderHint(pidFile, info)
		}

		return 0, daemonStarted, false, fmt.Errorf("pid file %s is locked by another process, but no daemon holds it after %s%s",
			clean.Log(pidFile), daemonPidTimeout, hint)
	}
}

// pidFileChanged reports whether the pid file described by after was created, replaced, or written since before.
func pidFileChanged(before, after os.FileInfo) bool {
	switch {
	case after == nil:
		return false
	case before == nil:
		return true
	}

	return !os.SameFile(before, after) || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime())
}

// waitForDaemonPid waits until the daemon has written its process ID to the pid file, which must have changed since
// before, and fails if the daemon exits first or the timeout passes.
func waitForDaemonPid(pidFile string, before os.FileInfo, child *os.Process, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if pid, info, err := readPidFile(pidFile); err == nil && pid == child.Pid && pidFileChanged(before, info) {
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

	if detachRequested(conf.DetachServer()) {
		conf.Shutdown()
		cancel()

		pid, status, detachErr := detachDaemon(dctx)

		switch {
		case detachErr != nil:
			return detachErr
		case status == daemonAlreadyRunning:
			log.Infof("daemon already running with process id %v", pid)
		default:
			log.Infof("daemon started with process id %v", pid)
		}

		return nil
	}

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
