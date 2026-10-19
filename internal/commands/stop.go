package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/user"
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

// processStart returns the session ID and start time of a process, see procStart.
var processStart = procStart

// probePidFileLock tests for a shared lock on an open pid file without waiting.
var probePidFileLock = func(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_SH|syscall.LOCK_NB) //nolint:gosec // File descriptors fit into an int.
}

// holderScanTimeout limits how long the stop command looks for the processes that have a pid file open.
var holderScanTimeout = 2 * time.Second

// scanHolders lists the processes that have a pid file open, see pidFileHolders.
var scanHolders = pidFileHolders

// lockStalePidFile takes the lock on a pid file before clearStalePidFile empties it.
var lockStalePidFile = func(lock *daemon.LockFile) error { return lock.Lock() }

// daemonPidFileFd is the file descriptor through which go-daemon passes the locked pid file to the daemon.
const daemonPidFileFd = 4

// pidFileMaxSize is the maximum size of a valid pid file in bytes.
const pidFileMaxSize = 32

// procClockTicks is the number of clock ticks per second in which /proc reports start times (USER_HZ).
const procClockTicks = 100

// startTimeTolerance allows for file systems that store modification times in steps of up to two seconds.
const startTimeTolerance = 2 * time.Second

// errDaemonUnconfirmed indicates that it cannot be confirmed whether a process holds the pid file.
var errDaemonUnconfirmed = errors.New("cannot confirm")

// errPidFileEmpty indicates a pid file without a process ID, as stop leaves it.
var errPidFileEmpty = errors.New("pid file is empty")

// errPidFileInvalid indicates a pid file whose content is not a valid process ID.
var errPidFileInvalid = errors.New("does not contain a valid process ID")

// errPidFileTooLarge indicates a pid file larger than pidFileMaxSize, which is not overwritten.
var errPidFileTooLarge = errors.New("is larger than 32 bytes")

// errPidFileLinked indicates a pid file with more than one name, which stop keeps and the daemon does not write.
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
	Description: "Sends the server that holds the PID file a termination signal, waits up to one minute until it has exited, and empties the PID file. " +
		"A PID file that no running server holds is out of date and is emptied instead, and an empty PID file means that no server is running. " +
		"Where the open files of the server cannot be read, e.g. by root without CAP_SYS_PTRACE, the PID file is kept, and if the server cannot be confirmed, stop reports how to proceed.",
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
// A pid file that no running daemon holds is out of date and emptied instead. Where the daemon was identified by
// its lock, the file is kept, and if the process that holds it cannot be confirmed, an error is returned.
func stopDaemon(pidFile string, timeout time.Duration) error {
	log.Infof("looking for pid in %s", clean.Log(pidFile))

	pid, info, err := readPidFile(pidFile)

	switch {
	case errors.Is(err, errPidFileEmpty):
		// A daemon that is starting holds the file before its process ID is written.
		if held, _ := pidFileLocked(pidFile, info); held {
			return fmt.Errorf("pid file %s is locked, but contains no process ID%s", clean.Log(pidFile), pidFileHolderHint(pidFile, info))
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

	held, viaLock, err := daemonHoldsPidFile(p, pidFile, info)

	if err != nil {
		return err
	} else if held {
		if err = p.Signal(syscall.SIGTERM); err != nil {
			return err
		}

		if err = waitForExit(p, timeout); err != nil {
			return err
		}

		log.Info("daemon exited successfully")

		// A daemon identified by its lock may have held another file, so the pid file is not changed.
		if viaLock {
			log.Infof("pid file %s was kept", clean.Log(pidFile))
			return nil
		}

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
// errPidFileEmpty, a larger file errPidFileTooLarge, and invalid content errPidFileInvalid, along with the file.
func readPidFile(pidFile string) (pid int, info os.FileInfo, err error) {
	f, info, err := openPidFile(pidFile, os.O_RDONLY)

	if err != nil {
		return 0, nil, err
	}

	defer func() { _ = f.Close() }()

	if info.Size() > pidFileMaxSize {
		return 0, info, fmt.Errorf("pid file %s %w", clean.Log(pidFile), errPidFileTooLarge)
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
		return 0, info, fmt.Errorf("pid file %s %w", clean.Log(pidFile), errPidFileInvalid)
	}

	return pid, info, nil
}

// daemonHoldsPidFile reports whether the process, named photoprism*, holds the pid file open, or the file it replaced.
// Where its open files cannot be read, the process must match the file (daemonCandidate) and the file must be locked,
// or errDaemonUnconfirmed is returned; viaLock reports that this fallback decided.
func daemonHoldsPidFile(p *os.Process, pidFile string, info os.FileInfo) (held, viaLock bool, err error) {
	if p == nil || p.Pid <= 0 || p.Pid == os.Getpid() || !daemonProcess(p) {
		return false, false, nil
	}

	if open, known := holdsFile(p.Pid, pidFile, info); known {
		return open, false, nil
	}

	var reason string

	mismatch := !daemonCandidate(p.Pid, info)

	if mismatch {
		reason = "it does not match the pid file"
	} else if locked, known := pidFileLocked(pidFile, info); !known {
		reason = "the lock of the pid file cannot be tested"
	} else if !locked {
		reason = "the pid file is not locked"
	} else {
		return true, true, nil
	}

	return false, true, fmt.Errorf("%w that process %d holds pid file %s, as %s%s",
		errDaemonUnconfirmed, p.Pid, clean.Log(pidFile), reason, unconfirmedRemedy(info, mismatch))
}

// unconfirmedRemedy suggests how to proceed where it cannot be confirmed that a process holds the pid file: run the
// command as the owner of the file, unless the caller is the owner, or remove the file if the process does not match.
func unconfirmedRemedy(info os.FileInfo, mismatch bool) string {
	const remove = "remove the pid file if no server is running"

	if info == nil {
		return ""
	}

	st, ok := info.Sys().(*syscall.Stat_t)

	if !ok {
		return ""
	} else if uint64(st.Uid) == uint64(os.Geteuid()) { //nolint:gosec // User IDs are not negative.
		return "; " + remove
	} else if mismatch {
		return fmt.Sprintf("; run the command as %s, who owns the pid file, or %s", fileOwnerName(info), remove)
	}

	return fmt.Sprintf("; run the command as %s, who owns the pid file", fileOwnerName(info))
}

// fileOwnerName describes the owner of the file as "user <name>", or as "uid <id>" if the name cannot be looked up.
func fileOwnerName(info os.FileInfo) string {
	if info == nil {
		return "its owner"
	}

	st, ok := info.Sys().(*syscall.Stat_t)

	if !ok {
		return "its owner"
	}

	uid := strconv.FormatUint(uint64(st.Uid), 10)

	if owner, err := user.LookupId(uid); err == nil && owner.Username != "" {
		return "user " + clean.Log(owner.Username)
	}

	return "uid " + uid
}

// processHasFile reports whether the process has the pid file described by info open under its name, or has the file
// open that the pid file replaced. known is false if its open files cannot be listed or read, e.g. by root without the
// capability to inspect processes of other users.
func processHasFile(pid int, pidFile string, info os.FileInfo) (open, known bool) {
	return processOpenFile(pid, info, replacedFileName(pidFile))
}

// processOpenFile reports whether the process has the file described by info open under its name, or, unless
// replaced is empty, the file that /proc lists as replaced. known is false if its open files cannot be read.
func processOpenFile(pid int, info os.FileInfo, replaced string) (open, known bool) {
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

		target, linkErr := readFdLink(fdName)

		// Compares the name first, which does not access the open file.
		switch {
		case errors.Is(linkErr, os.ErrPermission):
			return false, false
		case linkErr == nil && replaced != "" && target == replaced:
			if fdInfo, statErr := statFd(fdName); statErr == nil && fileReplaced(fdInfo, info) {
				return true, true
			} else if errors.Is(statErr, os.ErrPermission) {
				return false, false
			}

			continue
		case linkErr == nil && filepath.Base(target) != info.Name():
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

// replacedFileName returns the name under which /proc lists an open file that was at fileName before it was removed or
// replaced, or an empty string if the folder cannot be resolved.
func replacedFileName(fileName string) string {
	abs, err := filepath.Abs(fileName)

	if err != nil {
		return ""
	}

	dir, err := filepath.EvalSymlinks(filepath.Dir(abs))

	if err != nil {
		return ""
	}

	return filepath.Join(dir, filepath.Base(abs)) + " (deleted)"
}

// fileReplaced reports whether the open file has no name left and is on the same device as the file that replaced it.
func fileReplaced(open, info os.FileInfo) bool {
	if open == nil || info == nil {
		return false
	}

	st, ok := open.Sys().(*syscall.Stat_t)
	current, currentOk := info.Sys().(*syscall.Stat_t)

	return ok && currentOk && st.Nlink == 0 && st.Dev == current.Dev
}

// pidFileHolders returns the other processes that have the pid file described by info open, as "process <pid>
// (<name>)", and whether the open files of all processes could be read.
func pidFileHolders(info os.FileInfo) (holders []string, complete bool) {
	entries, err := os.ReadDir("/proc")

	if err != nil {
		return nil, false
	}

	complete = true

	for _, entry := range entries {
		pid, convErr := strconv.Atoi(entry.Name())

		if convErr != nil || pid <= 0 || pid == os.Getpid() {
			continue
		}

		// Only the current file can hold the lock, so a file it replaced is not matched.
		open, known := processOpenFile(pid, info, "")

		if !known {
			complete = false
			continue
		} else if !open {
			continue
		}

		cmdline, _ := procCmdline(pid)

		if name, _, _ := bytes.Cut(cmdline, []byte{0}); len(name) > 0 {
			holders = append(holders, fmt.Sprintf("process %d (%s)", pid, clean.Log(filepath.Base(string(name)))))
		} else {
			holders = append(holders, fmt.Sprintf("process %d", pid))
		}
	}

	return holders, complete
}

// pidFileHolderHint names the processes that have the pid file open, and how to find the process that holds it
// where none was found or not all could be listed within holderScanTimeout.
func pidFileHolderHint(pidFile string, info os.FileInfo) string {
	type scan struct {
		holders  []string
		complete bool
	}

	result := make(chan scan, 1)
	scanFunc := scanHolders

	go func() {
		holders, complete := scanFunc(info)
		result <- scan{holders, complete}
	}()

	var found scan

	select {
	case found = <-result:
	case <-time.After(holderScanTimeout):
	}

	hint := ""

	if len(found.holders) > 0 {
		hint = ", and is open in " + strings.Join(found.holders, ", ")
	}

	if !found.complete || len(found.holders) == 0 {
		hint += "; " + fuserHint(pidFile)
	}

	return hint
}

// fuserHint suggests running fuser or lsof on the pid file, naming it only if it is safe to paste in a shell.
func fuserHint(pidFile string) string {
	safe := func(r rune) bool {
		return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("._/-", r)
	}

	if pidFile != "" && !strings.HasPrefix(pidFile, "-") && strings.IndexFunc(pidFile, func(r rune) bool { return !safe(r) }) < 0 {
		return fmt.Sprintf("run fuser -v %[1]s or lsof %[1]s to find the process that holds it", pidFile)
	}

	return "run fuser -v or lsof on the pid file to find the process that holds it"
}

// pidFileLocked reports whether another process holds the exclusive lock on the pid file described by info, as a
// running daemon does. It tests for a shared lock, which a file opened read-only allows also on NFS. known is false
// if the lock cannot be tested, e.g. on platforms without file locks.
func pidFileLocked(pidFile string, info os.FileInfo) (held, known bool) {
	f, err := openSamePidFile(pidFile, os.O_RDONLY, info)

	if err != nil {
		return false, false
	}

	defer func() { _ = f.Close() }()

	switch err = probePidFileLock(f); {
	case err == nil:
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) //nolint:gosec // File descriptors fit into an int.
		return false, true
	case errors.Is(err, syscall.EWOULDBLOCK):
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

// daemonCandidate reports whether the process matches the pid file as the daemon would: it runs as the owner of the
// file, which has a single name and is not writable by its group or others, leads its own session, and started no
// later than the file was last written (startTimeTolerance). Without /proc, any process is accepted.
func daemonCandidate(pid int, info os.FileInfo) bool {
	if info == nil {
		return false
	}

	st, ok := info.Sys().(*syscall.Stat_t)

	if !ok {
		return true
	} else if st.Nlink > 1 || st.Mode&0o022 != 0 {
		return false
	}

	if uid, err := processUid(pid); err != nil {
		return !fs.FileExists("/proc/self/status")
	} else if uid != st.Uid {
		return false
	}

	session, started, err := processStart(pid)

	if err != nil {
		return !fs.FileExists("/proc/self/stat")
	}

	return session == pid && !started.After(info.ModTime().Add(startTimeTolerance))
}

// procStart returns the session ID and start time of the process from /proc, which can be read without the right
// to inspect the process.
func procStart(pid int) (session int, started time.Time, err error) {
	stat, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat")) //nolint:gosec // Reads /proc by process ID.

	if err != nil {
		return 0, time.Time{}, err
	}

	session, ticks, err := parseProcStat(stat)

	if err != nil {
		return 0, time.Time{}, err
	}

	system, err := os.ReadFile("/proc/stat")

	if err != nil {
		return 0, time.Time{}, err
	}

	boot, err := parseBootTime(system)

	if err != nil {
		return 0, time.Time{}, err
	}

	return session, boot.Add(ticksDuration(ticks)), nil
}

// ticksDuration converts clock ticks as reported by /proc into a duration.
func ticksDuration(ticks uint64) time.Duration {
	seconds, rest := ticks/procClockTicks, ticks%procClockTicks

	if seconds > math.MaxInt64/uint64(time.Second) {
		return time.Duration(math.MaxInt64)
	}

	return time.Duration(seconds)*time.Second + time.Duration(rest)*time.Second/procClockTicks //nolint:gosec // Bounded above.
}

// parseProcStat returns the session ID and the start time in clock ticks after boot from a /proc/<pid>/stat file,
// whose second field, the program name in parentheses, may contain spaces and parentheses.
func parseProcStat(stat []byte) (session int, startTicks uint64, err error) {
	end := bytes.LastIndexByte(stat, ')')

	if end < 0 {
		return 0, 0, errors.New("stat has no program name")
	}

	// The fields after the name start with the state (field 3); the session is field 6 and the start time field 22.
	fields := strings.Fields(string(stat[end+1:]))

	if len(fields) < 20 {
		return 0, 0, errors.New("stat has too few fields")
	}

	if session, err = strconv.Atoi(fields[3]); err != nil {
		return 0, 0, err
	}

	if startTicks, err = strconv.ParseUint(fields[19], 10, 64); err != nil {
		return 0, 0, err
	}

	return session, startTicks, nil
}

// parseBootTime returns the boot time from the btime line of /proc/stat.
func parseBootTime(stat []byte) (time.Time, error) {
	for _, line := range strings.Split(string(stat), "\n") {
		if fields := strings.Fields(line); len(fields) == 2 && fields[0] == "btime" {
			seconds, err := strconv.ParseInt(fields[1], 10, 64)
			return time.Unix(seconds, 0), err
		}
	}

	return time.Time{}, errors.New("stat has no boot time")
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
