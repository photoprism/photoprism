package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// stopTimeout limits how long the stop command waits for the daemon to exit.
const stopTimeout = time.Minute

// holdsFile reports whether a process has a file open, see processHasFile.
var holdsFile = processHasFile

// daemonPidFileFd is the file descriptor through which go-daemon passes the locked pid file to the daemon.
const daemonPidFileFd = 4

// closePidFileOnExec marks the pid file the daemon inherited as close-on-exec, so that programs it runs do not inherit it.
func closePidFileOnExec() {
	syscall.CloseOnExec(daemonPidFileFd)
}

// StopCommand configures the command name, flags, and action.
var StopCommand = &cli.Command{
	Name:    "stop",
	Aliases: []string{"down"},
	Usage:   "Stops the Web server running in daemon mode",
	Description: "Sends the server that holds the PID file a termination signal and waits up to one minute until it has exited. " +
		"A PID file that no running server holds is out of date and is removed instead.",
	Action: stopAction,
}

// stopAction stops the daemon if it is running.
func stopAction(ctx *cli.Context) error {
	conf, err := InitCoreConfig(ctx, false)

	if err != nil {
		log.Debug(err)
	}

	return stopDaemon(conf.PIDFilename(), stopTimeout)
}

// stopDaemon sends the daemon that holds the pid file a SIGTERM and waits until it has exited.
// A pid file that no running daemon holds is out of date and removed instead.
func stopDaemon(pidFile string, timeout time.Duration) error {
	log.Infof("looking for pid in %s", clean.Log(pidFile))

	pid, info, err := readPidFile(pidFile)

	if err != nil {
		if daemonNotRunning(err) {
			log.Info("daemon is not running")
			return nil
		}
		return err
	}

	p, err := os.FindProcess(pid)

	if err != nil {
		return err
	}

	if daemonHoldsPidFile(p, pidFile, info) {
		if err = p.Signal(syscall.SIGTERM); err != nil {
			return err
		}

		if err = waitForExit(p, timeout); err != nil {
			return err
		}

		log.Info("daemon exited successfully")

		return nil
	}

	switch err = removeStalePidFile(pidFile, info); {
	case err == nil:
		log.Infof("pid file named process %d, which is not a running PhotoPrism daemon, and was removed", pid)
		log.Info("daemon is not running")
		return nil
	case errors.Is(err, daemon.ErrWouldBlock):
		return fmt.Errorf("pid file %s is locked, and process %d is not a PhotoPrism daemon holding it", clean.Log(pidFile), pid)
	default:
		return err
	}
}

// openPidFile opens the pid file for reading without following a symbolic link or blocking, and requires a
// regular file. The caller closes it.
func openPidFile(pidFile string) (*os.File, os.FileInfo, error) {
	f, err := os.OpenFile(pidFile, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0) //nolint:gosec // Path from the config.

	if err != nil {
		if link, linkErr := os.Lstat(pidFile); linkErr == nil && link.Mode()&os.ModeSymlink != 0 { //nolint:gosec // Path from the config.
			return nil, nil, fmt.Errorf("pid file %s must not be a symbolic link", clean.Log(pidFile))
		}

		return nil, nil, err
	}

	info, err := f.Stat()

	if err == nil && !info.Mode().IsRegular() {
		err = fmt.Errorf("pid file %s is not a regular file", clean.Log(pidFile))
	}

	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}

	return f, info, nil
}

// openSamePidFile opens the pid file like openPidFile and requires it to be the file described by info.
func openSamePidFile(pidFile string, info os.FileInfo) (*os.File, error) {
	f, current, err := openPidFile(pidFile)

	if err != nil {
		return nil, err
	} else if info != nil && !os.SameFile(info, current) {
		_ = f.Close()
		return nil, fmt.Errorf("pid file %s was replaced", clean.Log(pidFile))
	}

	return f, nil
}

// readPidFile returns the process ID in the pid file and the file it was read from.
func readPidFile(pidFile string) (pid int, info os.FileInfo, err error) {
	f, info, err := openPidFile(pidFile)

	if err != nil {
		return 0, nil, err
	}

	defer func() { _ = f.Close() }()

	b, err := io.ReadAll(io.LimitReader(f, 32))

	if err != nil {
		return 0, nil, err
	}

	if pid, err = strconv.Atoi(strings.TrimSpace(string(b))); err != nil || pid <= 0 || pid > math.MaxInt32 {
		return 0, nil, fmt.Errorf("pid file %s does not contain a valid process ID", clean.Log(pidFile))
	}

	return pid, info, nil
}

// daemonHoldsPidFile reports whether the process holds the pid file open, as the daemon does, and is named
// photoprism*. Where its open files cannot be listed, the pid file must be locked instead.
func daemonHoldsPidFile(p *os.Process, pidFile string, info os.FileInfo) bool {
	if p == nil || p.Pid <= 0 || p.Pid == os.Getpid() || !daemonProcess(p) {
		return false
	}

	if open, known := holdsFile(p.Pid, info); known {
		return open
	}

	held, known := pidFileLocked(pidFile, info)

	return held || !known
}

// processHasFile reports whether the process has the file open under its name. known is false if its open files cannot be
// listed.
func processHasFile(pid int, info os.FileInfo) (open, known bool) {
	if info == nil {
		return false, false
	}

	dir := filepath.Join("/proc", strconv.Itoa(pid), "fd")
	entries, err := os.ReadDir(dir)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) && fs.PathExists("/proc/self/fd") {
			return false, true
		}

		return false, false
	}

	for _, entry := range entries {
		fdName := filepath.Join(dir, entry.Name())

		// Compares the name first, which does not access the open file.
		if target, linkErr := os.Readlink(fdName); linkErr == nil && filepath.Base(target) != info.Name() {
			continue
		}

		if fdInfo, statErr := os.Stat(fdName); statErr == nil && os.SameFile(info, fdInfo) { //nolint:gosec // Reads /proc.
			return true, true
		}
	}

	return false, true
}

// pidFileLocked reports whether another process holds the lock on the pid file described by info, as a running
// daemon does. known is false if the lock cannot be tested, e.g. on platforms without file locks.
func pidFileLocked(pidFile string, info os.FileInfo) (held, known bool) {
	f, err := openSamePidFile(pidFile, info)

	if err != nil {
		return false, false
	}

	lock := daemon.NewLockFile(f)

	defer func() { _ = lock.Close() }()

	if err = lock.Lock(); err == nil {
		_ = lock.Unlock()
		return false, true
	} else if errors.Is(err, daemon.ErrWouldBlock) {
		return true, true
	}

	return false, false
}

// removeStalePidFile removes the pid file described by info while holding its lock, and returns daemon.ErrWouldBlock
// if another process holds it. Where the lock cannot be tested, the file is removed.
func removeStalePidFile(pidFile string, info os.FileInfo) error {
	f, err := openSamePidFile(pidFile, info)

	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	lock := daemon.NewLockFile(f)

	defer func() { _ = lock.Close() }()

	if err = lock.Lock(); errors.Is(err, daemon.ErrWouldBlock) {
		return err
	}

	if err = os.Remove(pidFile); err != nil && !errors.Is(err, os.ErrNotExist) { //nolint:gosec // Path from the config.
		return err
	}

	return nil
}

// daemonProcess reports whether the process runs and its program name starts with "photoprism".
// Without /proc, the name cannot be read, so any running process is accepted.
func daemonProcess(p *os.Process) bool {
	if p == nil {
		return false
	}

	cmdline, err := procCmdline(p.Pid)

	if err != nil {
		if fs.FileExists("/proc/self/cmdline") {
			return false
		}

		return p.Signal(syscall.Signal(0)) == nil
	}

	name, _, _ := bytes.Cut(cmdline, []byte{0})

	return strings.HasPrefix(filepath.Base(string(name)), "photoprism")
}

// processRunning reports whether the process runs, counting one that has exited but was not reaped yet as gone.
func processRunning(p *os.Process) bool {
	if p == nil || p.Signal(syscall.Signal(0)) != nil {
		return false
	}

	cmdline, err := procCmdline(p.Pid)

	if err != nil {
		return !fs.FileExists("/proc/self/cmdline")
	}

	return len(cmdline) > 0
}

// procCmdline returns the command line of the process from /proc, with its arguments separated by NUL bytes.
func procCmdline(pid int) ([]byte, error) {
	return os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline")) //nolint:gosec // Reads /proc by process ID.
}

// waitForExit polls until the process has exited, and fails after the timeout.
func waitForExit(p *os.Process, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for processRunning(p) {
		if time.Now().After(deadline) {
			return fmt.Errorf("daemon %d has not exited within %s", p.Pid, timeout)
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// daemonNotRunning reports whether a failed pid file read means that no daemon is running.
func daemonNotRunning(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
