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

// clearTimeout limits how long the stop command waits for the lock of a daemon that has exited.
const clearTimeout = time.Second

// holdsFile reports whether a process has a file open, see processHasFile.
var holdsFile = processHasFile

// readFdLink reads the target of an open file listed in /proc.
var readFdLink = os.Readlink

// statFd returns the information of an open file listed in /proc.
var statFd = os.Stat

// processUid returns the effective user ID of a process, see procUid.
var processUid = procUid

// lockStalePidFile takes the lock on a pid file before clearStalePidFile empties it.
var lockStalePidFile = func(lock *daemon.LockFile) error { return lock.Lock() }

// daemonPidFileFd is the file descriptor through which go-daemon passes the locked pid file to the daemon.
const daemonPidFileFd = 4

// pidFileMaxSize is the maximum size of a valid pid file in bytes.
const pidFileMaxSize = 32

// errPidFileEmpty indicates a pid file without a process ID, as stop leaves it.
var errPidFileEmpty = errors.New("pid file is empty")

// errPidFileLinked indicates a pid file with more than one name, which stop keeps.
var errPidFileLinked = errors.New("pid file has more than one link")

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
		"A PID file that no running server holds is out of date and is emptied instead, and an empty PID file means that no server is running.",
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

// stopDaemon sends the daemon that holds the pid file a SIGTERM, waits until it has exited, and empties the file.
// A pid file that no running daemon holds is out of date and emptied instead.
func stopDaemon(pidFile string, timeout time.Duration) error {
	log.Infof("looking for pid in %s", clean.Log(pidFile))

	pid, info, err := readPidFile(pidFile)

	switch {
	case errors.Is(err, errPidFileEmpty):
		// A daemon that is starting holds the file before its process ID is written.
		if held, _ := pidFileLocked(pidFile, info); held {
			return fmt.Errorf("pid file %s is locked, but contains no process ID", clean.Log(pidFile))
		}

		log.Info("daemon is not running")
		return nil
	case daemonNotRunning(err):
		log.Info("daemon is not running")
		return nil
	case err != nil:
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

		// A daemon that has started meanwhile holds the lock, so its file is kept.
		if err = clearExitedPidFile(pidFile, info, clearTimeout); errors.Is(err, daemon.ErrWouldBlock) {
			log.Infof("pid file %s is locked by another process and was kept", clean.Log(pidFile))
		} else if err != nil {
			log.Warnf("pid file %s was not emptied (%s)", clean.Log(pidFile), clean.Error(err))
		}

		return nil
	}

	switch err = clearStalePidFile(pidFile, info); {
	case err == nil:
		log.Infof("pid file named process %d, which is not a running PhotoPrism daemon, and was emptied", pid)
		log.Info("daemon is not running")
		return nil
	case errors.Is(err, errPidFileLinked):
		// The process ID it keeps is checked again on the next run.
		log.Warnf("pid file %s has more than one link and was kept", clean.Log(pidFile))
		log.Info("daemon is not running")
		return nil
	case errors.Is(err, daemon.ErrWouldBlock):
		return fmt.Errorf("pid file %s is locked, and process %d it named is not a PhotoPrism daemon holding it", clean.Log(pidFile), pid)
	default:
		return err
	}
}

// openPidFile opens the pid file with the access mode in flag, e.g. os.O_RDONLY, without following a symbolic link,
// blocking, or creating it, and requires a regular file. The caller closes it.
func openPidFile(pidFile string, flag int) (*os.File, os.FileInfo, error) {
	f, err := os.OpenFile(pidFile, flag&syscall.O_ACCMODE|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0) //nolint:gosec // Path from the config.

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
func openSamePidFile(pidFile string, flag int, info os.FileInfo) (*os.File, error) {
	f, current, err := openPidFile(pidFile, flag)

	if err != nil {
		return nil, err
	} else if info != nil && !os.SameFile(info, current) {
		_ = f.Close()
		return nil, fmt.Errorf("pid file %s was replaced", clean.Log(pidFile))
	}

	return f, nil
}

// readPidFile returns the process ID in the pid file and the file it was read from. An empty file returns
// errPidFileEmpty along with the file.
func readPidFile(pidFile string) (pid int, info os.FileInfo, err error) {
	f, info, err := openPidFile(pidFile, os.O_RDONLY)

	if err != nil {
		return 0, nil, err
	}

	defer func() { _ = f.Close() }()

	if info.Size() > pidFileMaxSize {
		return 0, nil, fmt.Errorf("pid file %s does not contain a valid process ID", clean.Log(pidFile))
	}

	b, err := io.ReadAll(io.LimitReader(f, pidFileMaxSize))

	if err != nil {
		return 0, nil, err
	}

	content := strings.TrimSpace(string(b))

	if content == "" {
		return 0, info, fmt.Errorf("%s: %w", clean.Log(pidFile), errPidFileEmpty)
	}

	if pid, err = strconv.Atoi(content); err != nil || pid <= 0 || pid > math.MaxInt32 {
		return 0, nil, fmt.Errorf("pid file %s does not contain a valid process ID", clean.Log(pidFile))
	}

	return pid, info, nil
}

// daemonHoldsPidFile reports whether the process holds the pid file open, as the daemon does, and is named
// photoprism*. Where its open files cannot be listed or read, the pid file must be locked instead, by any process,
// and the process must run as the owner of the file.
func daemonHoldsPidFile(p *os.Process, pidFile string, info os.FileInfo) bool {
	if p == nil || p.Pid <= 0 || p.Pid == os.Getpid() || !daemonProcess(p) {
		return false
	}

	if open, known := holdsFile(p.Pid, info); known {
		return open
	} else if !processOwnsFile(p.Pid, info) {
		log.Debugf("cannot confirm that process %d holds pid file %s", p.Pid, clean.Log(pidFile))
		return false
	}

	held, known := pidFileLocked(pidFile, info)

	return held || !known
}

// processHasFile reports whether the process has the file open under its name. known is false if its open files cannot be
// listed or read, e.g. by root without the capability to inspect processes of other users.
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
		if target, linkErr := readFdLink(fdName); errors.Is(linkErr, os.ErrPermission) {
			return false, false
		} else if linkErr == nil && filepath.Base(target) != info.Name() {
			continue
		}

		if fdInfo, statErr := statFd(fdName); statErr == nil && os.SameFile(info, fdInfo) {
			return true, true
		} else if errors.Is(statErr, os.ErrPermission) {
			return false, false
		}
	}

	return false, true
}

// pidFileLocked reports whether another process holds the lock on the pid file described by info, as a running
// daemon does. known is false if the lock cannot be tested, e.g. on platforms without file locks.
func pidFileLocked(pidFile string, info os.FileInfo) (held, known bool) {
	f, err := openSamePidFile(pidFile, os.O_RDONLY, info)

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

// clearStalePidFile empties the pid file described by info while holding its lock, and returns daemon.ErrWouldBlock
// if another process holds it. Where the lock cannot be tested, the file is emptied. A file with more than one
// name is kept (errPidFileLinked), and the file is never removed.
func clearStalePidFile(pidFile string, info os.FileInfo) error {
	f, err := openSamePidFile(pidFile, os.O_WRONLY, info)

	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	lock := daemon.NewLockFile(f)

	defer func() { _ = lock.Close() }()

	if err = lockStalePidFile(lock); errors.Is(err, daemon.ErrWouldBlock) {
		return err
	}

	// Checks under the lock that the path still names the locked file, and that it has no other name.
	if locked, statErr := f.Stat(); statErr != nil {
		return statErr
	} else if current, lstatErr := os.Lstat(pidFile); lstatErr != nil || !os.SameFile(locked, current) { //nolint:gosec // Path from the config.
		return fmt.Errorf("pid file %s was replaced", clean.Log(pidFile))
	} else if st, ok := locked.Sys().(*syscall.Stat_t); ok && st.Nlink > 1 {
		return fmt.Errorf("%s: %w", clean.Log(pidFile), errPidFileLinked)
	}

	return f.Truncate(0)
}

// clearExitedPidFile empties the pid file of a daemon that has exited, like clearStalePidFile, and retries until
// the timeout while its lock is held, as the files of an exiting process close after its command line is gone.
func clearExitedPidFile(pidFile string, info os.FileInfo, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		err := clearStalePidFile(pidFile, info)

		if !errors.Is(err, daemon.ErrWouldBlock) || time.Now().After(deadline) {
			return err
		}

		time.Sleep(20 * time.Millisecond)
	}
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

// processOwnsFile reports whether the process runs as the owner of the file, which must have a single name and must
// not be writable by its group or others. Without /proc, the user cannot be read, so any process is accepted.
func processOwnsFile(pid int, info os.FileInfo) bool {
	if info == nil {
		return false
	}

	st, ok := info.Sys().(*syscall.Stat_t)

	if !ok {
		return true
	} else if st.Nlink > 1 || st.Mode&0o022 != 0 {
		return false
	}

	uid, err := processUid(pid)

	if err != nil {
		return !fs.FileExists("/proc/self/status")
	}

	return uid == st.Uid
}

// procUid returns the effective user ID of the process from /proc, which can be read without the right to
// inspect the process.
func procUid(pid int) (uint32, error) {
	b, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "status")) //nolint:gosec // Reads /proc by process ID.

	if err != nil {
		return 0, err
	}

	return statusUid(b)
}

// statusUid returns the effective user ID from the Uid line of a /proc status file, which lists the real,
// effective, saved, and file system user IDs.
func statusUid(status []byte) (uint32, error) {
	for _, line := range strings.Split(string(status), "\n") {
		if fields := strings.Fields(line); len(fields) >= 3 && fields[0] == "Uid:" {
			uid, err := strconv.ParseUint(fields[2], 10, 32)
			return uint32(uid), err
		}
	}

	return 0, errors.New("status has no user ID")
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
