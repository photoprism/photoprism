package commands

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/sevlyar/go-daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestDaemonNotRunning covers which pid file read failures mean that no daemon is running.
func TestDaemonNotRunning(t *testing.T) {
	readErr := func(t *testing.T, pidFile string) error {
		_, _, err := readPidFile(pidFile)
		require.Error(t, err)
		return err
	}

	t.Run("NoPidFile", func(t *testing.T) {
		assert.True(t, daemonNotRunning(readErr(t, filepath.Join(t.TempDir(), "photoprism.pid"))))
	})
	t.Run("InvalidPidFile", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.WriteFile(pidFile, []byte("not a pid"), fs.ModeFile))
		assert.False(t, daemonNotRunning(readErr(t, pidFile)))
	})
	t.Run("DanglingSymlink", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.Symlink(filepath.Join(t.TempDir(), "missing.pid"), pidFile))
		assert.False(t, daemonNotRunning(readErr(t, pidFile)))
	})
	t.Run("Directory", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.Mkdir(pidFile, fs.ModeDir))
		assert.False(t, daemonNotRunning(readErr(t, pidFile)))
	})
}

// deadPid is a process ID above the highest that the kernel assigns, so that it never names a process.
var deadPid = unusedPid()

// deadPidText is deadPid as it appears in a pid file.
var deadPidText = strconv.Itoa(deadPid)

// unusedPid returns a process ID above /proc/sys/kernel/pid_max, or the largest valid one where it cannot be read.
func unusedPid() int {
	if b, err := os.ReadFile("/proc/sys/kernel/pid_max"); err == nil {
		if pidMax, parseErr := strconv.Atoi(strings.TrimSpace(string(b))); parseErr == nil && pidMax < math.MaxInt32 {
			return pidMax + 1
		}
	}

	return math.MaxInt32
}

// TestUnusedPid covers choosing a process ID that never names a process.
func TestUnusedPid(t *testing.T) {
	requireProc(t)

	b, err := os.ReadFile("/proc/sys/kernel/pid_max")
	require.NoError(t, err)
	pidMax, err := strconv.Atoi(strings.TrimSpace(string(b)))
	require.NoError(t, err)

	assert.Greater(t, deadPid, pidMax)
	assert.LessOrEqual(t, deadPid, math.MaxInt32)

	pid, _, err := readPidFile(unlockedPidFile(t, deadPid))
	require.NoError(t, err)
	assert.Equal(t, deadPid, pid)
	assert.Error(t, syscall.Kill(deadPid, 0))
}

// stopHelperEnv selects the stand-in process the test binary runs as for the stop tests.
const stopHelperEnv = "PHOTOPRISM_TEST_STOP_HELPER"

// stopHelperPidEnv names the pid file of the stop-self and reborn stand-ins.
const stopHelperPidEnv = "PHOTOPRISM_TEST_STOP_PID_FILE"

// stopHelperLogEnv names the file to which the reborn stand-in's daemon writes its report.
const stopHelperLogEnv = "PHOTOPRISM_TEST_STOP_LOG_FILE"

// runStopHelper runs the test binary as a stand-in process for the stop tests and returns its exit code.
func runStopHelper(mode string) int {
	switch mode {
	case "ignore-term":
		signal.Ignore(syscall.SIGTERM)
		fmt.Println("ready")
		time.Sleep(time.Minute)
	case "subprocess", "subprocess-cloexec":
		// Holds the pid file as fd 4 like the daemon, runs a program and prints its process ID.
		if mode == "subprocess-cloexec" {
			closePidFileOnExec()
		}

		cmd := exec.Command("sleep", "30") //nolint:gosec // Runs sleep from PATH with a fixed duration.

		if err := cmd.Start(); err != nil {
			fmt.Println(err)
			return 1
		}

		fmt.Println(cmd.Process.Pid)
		_ = cmd.Wait()
	case "prologue":
		// Holds a file as fd 4 without having been started by go-daemon.
		before, _ := unix.FcntlInt(daemonPidFileFd, unix.F_GETFD, 0)
		err := startPrologue()
		after, _ := unix.FcntlInt(daemonPidFileFd, unix.F_GETFD, 0)
		fmt.Printf("cloexec-before=%t cloexec-after=%t err=%v\n", before&unix.FD_CLOEXEC != 0, after&unix.FD_CLOEXEC != 0, err)
	case "write-pid", "write-pid-short":
		// Runs the prologue as a daemon that inherited fd 4, see TestWriteDaemonPid.
		if mode == "write-pid-short" {
			// Writes one byte, after checking that the file was emptied first.
			writePidFd = func(fd int, b []byte) (int, error) {
				var st syscall.Stat_t

				if err := syscall.Fstat(fd, &st); err != nil {
					return 0, err
				} else if st.Size != 0 {
					return 0, fmt.Errorf("file has %d bytes before writing", st.Size)
				}

				return syscall.Write(fd, b[:1])
			}
		}

		fmt.Printf("pid=%d err=%v\n", os.Getpid(), startPrologue())
	case "reborn":
		return runRebornHelper(os.Getenv(stopHelperPidEnv), os.Getenv(stopHelperLogEnv))
	case "stop-self":
		pidFile := os.Getenv(stopHelperPidEnv)
		lock, err := daemon.CreatePidFile(pidFile, fs.ModeFile)

		if err != nil {
			fmt.Println(err)
			return 1
		}

		defer func() { _ = lock.Close() }()

		fmt.Println(stopDaemon(pidFile, time.Second))
	default:
		return 2
	}

	return 0
}

// runRebornHelper starts a daemon with go-daemon like the start command and prints its process ID. The daemon
// writes whether fd 4 is the pid file and whether it is close-on-exec before and after startPrologue to its log.
func runRebornHelper(pidFile, logFile string) int {
	if !daemon.WasReborn() {
		dctx := &daemon.Context{PidFileName: pidFile, LogFileName: logFile, Args: []string{"photoprism-test", "-test.run=^$"}}

		// Starts the daemon and waits for its process ID, like the start command.
		pid, _, err := detachDaemon(dctx)

		if err != nil {
			fmt.Println(err)
			return 1
		}

		fmt.Println(pid)
		return 0
	}

	fdInfo, fdErr := os.Stat(fmt.Sprintf("/proc/self/fd/%d", daemonPidFileFd))
	pidInfo, pidErr := os.Stat(pidFile) //nolint:gosec // Path from the test.
	before, _ := unix.FcntlInt(daemonPidFileFd, unix.F_GETFD, 0)
	prologueErr := startPrologue()
	after, _ := unix.FcntlInt(daemonPidFileFd, unix.F_GETFD, 0)

	fmt.Printf("pid-file=%t cloexec-before=%t cloexec-after=%t err=%v\n", fdErr == nil && pidErr == nil && os.SameFile(fdInfo, pidInfo),
		before&unix.FD_CLOEXEC != 0, after&unix.FD_CLOEXEC != 0, prologueErr)

	// Keeps running like a daemon until the test stops it.
	time.Sleep(time.Minute)

	return 0
}

// requireProc skips the test where /proc, which identifies the stand-in processes, is not available.
func requireProc(t *testing.T) {
	t.Helper()

	if !fs.PathExists("/proc/self/fd") {
		t.Skip("requires /proc")
	}
}

// lockedPidFile creates a locked, empty pid file and returns its name and the open file a daemon would hold.
func lockedPidFile(t *testing.T) (string, *os.File) {
	t.Helper()

	pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
	lock, err := daemon.OpenLockFile(pidFile, daemon.FILE_PERM)
	require.NoError(t, err)
	require.NoError(t, lock.Lock())
	t.Cleanup(func() { _ = lock.Close() })

	return pidFile, lock.File
}

// sharedLock takes a shared lock on the file, as a process that reads it may, and returns the open file.
func sharedLock(t *testing.T, fileName string) *os.File {
	t.Helper()

	f, err := os.Open(fileName) //nolint:gosec // Opens a test file.
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	require.NoError(t, syscall.Flock(int(f.Fd()), syscall.LOCK_SH|syscall.LOCK_NB)) //nolint:gosec // File descriptors fit into an int.

	return f
}

// unlockedPidFile writes a pid file for the process that no process holds.
func unlockedPidFile(t *testing.T, pid int) string {
	t.Helper()

	pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
	require.NoError(t, os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), daemon.FILE_PERM))

	return pidFile
}

// startProcess runs sleep with the given program name in argv[0], passes it the files to hold open,
// and stops it when the test ends.
func startProcess(t *testing.T, name, seconds string, files ...*os.File) *exec.Cmd {
	t.Helper()
	requireProc(t)

	sleep, err := exec.LookPath("sleep")
	require.NoError(t, err)

	cmd := exec.Command(sleep, seconds) //nolint:gosec // Runs sleep from PATH with a fixed duration.
	cmd.Args[0] = name
	cmd.ExtraFiles = files
	// Leads its own session, like a daemon.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	require.NoError(t, cmd.Start())

	done := make(chan struct{})

	go func() {
		_ = cmd.Wait()
		close(done)
	}()

	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		<-done
	})

	waitForName(t, cmd.Process.Pid, name)

	return cmd
}

// startDaemon runs a stand-in daemon named photoprism-test that holds a locked pid file naming it.
func startDaemon(t *testing.T, seconds string) (*exec.Cmd, string) {
	t.Helper()

	pidFile, f := lockedPidFile(t)
	cmd := startProcess(t, "photoprism-test", seconds, f)
	// Leaves the lock to the stand-in, like the parent of a real daemon.
	require.NoError(t, f.Close())
	require.NoError(t, fs.WriteString(pidFile, strconv.Itoa(cmd.Process.Pid)))

	return cmd, pidFile
}

// startHelper runs the test binary as a stand-in process named photoprism-test in the given helper mode,
// with the additional environment variables and the files to hold open.
func startHelper(t *testing.T, mode string, env []string, files ...*os.File) (*exec.Cmd, *bufio.Reader) {
	t.Helper()

	return startHelperInput(t, mode, env, nil, files...)
}

// startHelperInput runs the test binary as a stand-in like startHelper and passes it stdin as its standard input.
func startHelperInput(t *testing.T, mode string, env []string, stdin io.Reader, files ...*os.File) (*exec.Cmd, *bufio.Reader) {
	t.Helper()
	requireProc(t)

	cmd := exec.Command(os.Args[0], "-test.run=^$") //nolint:gosec // Runs this test binary as a stand-in.
	cmd.Args[0] = "photoprism-test"
	cmd.Stdin = stdin
	cmd.Env = append(append(os.Environ(), stopHelperEnv+"="+mode), env...)
	cmd.ExtraFiles = files
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())

	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	return cmd, bufio.NewReader(stdout)
}

// waitForName waits until the process runs the program with the given name in argv[0].
func waitForName(t *testing.T, pid int, name string) {
	t.Helper()

	cmdline := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")

	require.Eventually(t, func() bool {
		b, err := os.ReadFile(cmdline) //nolint:gosec // Reads the stand-in process command line.
		return err == nil && strings.HasPrefix(string(b), name+"\x00")
	}, 5*time.Second, 10*time.Millisecond)
}

// replacePidFile writes the content to a new file and renames it over the pid file, as tools that replace files do.
func replacePidFile(t *testing.T, pidFile, content string) {
	t.Helper()

	require.NoError(t, os.WriteFile(pidFile+".new", []byte(content), daemon.FILE_PERM))
	require.NoError(t, os.Rename(pidFile+".new", pidFile))
}

// fileInfo returns the information of an existing file.
func fileInfo(t *testing.T, fileName string) os.FileInfo {
	t.Helper()

	info, err := os.Stat(fileName)
	require.NoError(t, err)

	return info
}

// keepFile holds the file open until the test ends, so that its inode number cannot be reused.
func keepFile(t *testing.T, fileName string) os.FileInfo {
	t.Helper()

	f, err := os.Open(fileName) //nolint:gosec // Opens a test file.
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	info, err := f.Stat()
	require.NoError(t, err)

	return info
}

// assertPidFileContent checks that the pid file exists and has the content.
func assertPidFileContent(t *testing.T, pidFile, content string) {
	t.Helper()

	b, err := os.ReadFile(pidFile) //nolint:gosec // Reads a test file.
	require.NoError(t, err)
	assert.Equal(t, content, string(b))
}

// assertEmptyPidFile checks that the pid file exists and is empty.
func assertEmptyPidFile(t *testing.T, pidFile string) {
	t.Helper()
	assertPidFileContent(t, pidFile, "")
}

// TestReadPidFile covers reading the process ID from a pid file.
func TestReadPidFile(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 1234)
		pid, info, err := readPidFile(pidFile)
		require.NoError(t, err)
		assert.Equal(t, 1234, pid)
		assert.True(t, os.SameFile(fileInfo(t, pidFile), info))
	})
	t.Run("Symlink", func(t *testing.T) {
		target := unlockedPidFile(t, 1234)
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.Symlink(target, pidFile))
		_, _, err := readPidFile(pidFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be a symbolic link")
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, content := range []string{"abc", "0", "-1", "-1234", "4294967297"} {
			pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
			require.NoError(t, os.WriteFile(pidFile, []byte(content), fs.ModeFile))
			_, info, err := readPidFile(pidFile)
			assert.ErrorIs(t, err, errPidFileInvalid, content)
			assert.NotErrorIs(t, err, errPidFileEmpty, content)
			assert.NotNil(t, info, content)
		}
	})
	t.Run("MaxSize", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.WriteFile(pidFile, []byte(strings.Repeat(" ", pidFileMaxSize-4)+"1234"), fs.ModeFile))
		pid, _, err := readPidFile(pidFile)
		require.NoError(t, err)
		assert.Equal(t, 1234, pid)
	})
	t.Run("Oversized", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.WriteFile(pidFile, []byte(strings.Repeat(" ", pidFileMaxSize)+"1234"), fs.ModeFile))
		_, info, err := readPidFile(pidFile)
		assert.ErrorIs(t, err, errPidFileTooLarge)
		assert.NotErrorIs(t, err, errPidFileInvalid)
		assert.NotErrorIs(t, err, errPidFileEmpty)
		assert.NotNil(t, info)
		assert.Contains(t, err.Error(), "is larger than 32 bytes")
	})
	t.Run("Empty", func(t *testing.T) {
		for _, content := range []string{"", "\n", " \t\n"} {
			pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
			require.NoError(t, os.WriteFile(pidFile, []byte(content), fs.ModeFile))
			pid, info, err := readPidFile(pidFile)
			assert.ErrorIs(t, err, errPidFileEmpty, content)
			assert.Equal(t, 0, pid)
			require.NotNil(t, info)
			assert.True(t, os.SameFile(fileInfo(t, pidFile), info))
		}
	})
	t.Run("Missing", func(t *testing.T) {
		_, _, err := readPidFile(filepath.Join(t.TempDir(), "photoprism.pid"))
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
	t.Run("Fifo", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, syscall.Mkfifo(pidFile, 0o600))
		done := make(chan error, 1)

		go func() {
			_, _, err := readPidFile(pidFile)
			done <- err
		}()

		select {
		case err := <-done:
			require.Error(t, err)
			assert.Contains(t, err.Error(), "is not a regular file")
		case <-time.After(5 * time.Second):
			t.Fatal("reading a FIFO blocked")
		}
	})
}

// TestDaemonProcess covers which processes are recognized by their program name.
func TestDaemonProcess(t *testing.T) {
	t.Run("PhotoPrism", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		assert.True(t, daemonProcess(cmd.Process))
	})
	t.Run("FullPath", func(t *testing.T) {
		cmd := startProcess(t, "/opt/photoprism/bin/photoprism", "30")
		assert.True(t, daemonProcess(cmd.Process))
	})
	t.Run("OtherProgram", func(t *testing.T) {
		cmd := startProcess(t, "/usr/bin/sleep", "30")
		assert.False(t, daemonProcess(cmd.Process))
	})
	t.Run("NotRunning", func(t *testing.T) {
		requireProc(t)
		p, err := os.FindProcess(deadPid)
		require.NoError(t, err)
		assert.False(t, daemonProcess(p))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, daemonProcess(nil))
	})
}

// TestProcessHasFile covers detecting whether a process holds a file open.
func TestProcessHasFile(t *testing.T) {
	t.Run("Held", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		open, known := processHasFile(cmd.Process.Pid, pidFile, fileInfo(t, pidFile))
		assert.True(t, known)
		assert.True(t, open)
	})
	t.Run("NotHeld", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		pidFile := unlockedPidFile(t, cmd.Process.Pid)
		open, known := processHasFile(cmd.Process.Pid, pidFile, fileInfo(t, pidFile))
		assert.True(t, known)
		assert.False(t, open)
	})
	t.Run("NotRunning", func(t *testing.T) {
		requireProc(t)
		pidFile := unlockedPidFile(t, deadPid)
		open, known := processHasFile(deadPid, pidFile, fileInfo(t, pidFile))
		assert.True(t, known)
		assert.False(t, open)
	})
	t.Run("NoFile", func(t *testing.T) {
		_, known := processHasFile(os.Getpid(), filepath.Join(t.TempDir(), "photoprism.pid"), nil)
		assert.False(t, known)
	})
	t.Run("Replaced", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		replacePidFile(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		open, known := processHasFile(cmd.Process.Pid, pidFile, fileInfo(t, pidFile))
		assert.True(t, known)
		assert.True(t, open)
	})
	t.Run("ReplacedSymlinkedFolder", func(t *testing.T) {
		cmd, target := startDaemon(t, "30")
		dir := filepath.Join(t.TempDir(), "storage")
		require.NoError(t, os.Symlink(filepath.Dir(target), dir))
		pidFile := filepath.Join(dir, filepath.Base(target))
		replacePidFile(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		open, known := processHasFile(cmd.Process.Pid, pidFile, fileInfo(t, pidFile))
		assert.True(t, known)
		assert.True(t, open)
	})
	t.Run("ReplacedOtherFolder", func(t *testing.T) {
		cmd, held := startDaemon(t, "30")
		replacePidFile(t, held, strconv.Itoa(cmd.Process.Pid))
		pidFile := unlockedPidFile(t, cmd.Process.Pid)
		require.Equal(t, filepath.Base(held), filepath.Base(pidFile))
		open, known := processHasFile(cmd.Process.Pid, pidFile, fileInfo(t, pidFile))
		assert.True(t, known)
		assert.False(t, open)
	})
	t.Run("NamedLikeReplaced", func(t *testing.T) {
		dir := t.TempDir()
		pidFile := filepath.Join(dir, "photoprism.pid")
		f, err := os.Create(pidFile + " (deleted)") //nolint:gosec // Creates a test file.
		require.NoError(t, err)
		t.Cleanup(func() { _ = f.Close() })
		cmd := startProcess(t, "photoprism-test", "30", f)
		require.NoError(t, os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), daemon.FILE_PERM))
		open, known := processHasFile(cmd.Process.Pid, pidFile, fileInfo(t, pidFile))
		assert.True(t, known)
		assert.False(t, open)
	})
	t.Run("HardLink", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		other := filepath.Join(t.TempDir(), "other.pid")
		require.NoError(t, os.Link(pidFile, other))
		open, known := processHasFile(cmd.Process.Pid, other, fileInfo(t, other))
		assert.True(t, known)
		assert.False(t, open)
		// The stand-in holds the lock, so stop refuses rather than emptying the file.
		err := stopDaemon(other, time.Second)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "is not a PhotoPrism daemon holding it")
		assertPidFileContent(t, other, strconv.Itoa(cmd.Process.Pid))
		assert.True(t, processRunning(cmd.Process))
	})
	t.Run("PermissionDenied", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")

		defaultReadFdLink := readFdLink
		readFdLink = func(name string) (string, error) {
			return "", &os.PathError{Op: "readlink", Path: name, Err: syscall.EACCES}
		}
		t.Cleanup(func() { readFdLink = defaultReadFdLink })

		open, known := processHasFile(cmd.Process.Pid, pidFile, fileInfo(t, pidFile))
		assert.False(t, known)
		assert.False(t, open)
		// Falls back to the lock, which the stand-in holds.
		assert.True(t, daemonHolds(t, cmd.Process, pidFile, fileInfo(t, pidFile)))
	})
	t.Run("StatPermissionDenied", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")

		defaultStatFd := statFd
		statFd = func(name string) (os.FileInfo, error) {
			return nil, &os.PathError{Op: "stat", Path: name, Err: syscall.EACCES}
		}
		t.Cleanup(func() { statFd = defaultStatFd })

		open, known := processHasFile(cmd.Process.Pid, pidFile, fileInfo(t, pidFile))
		assert.False(t, known)
		assert.False(t, open)
	})
	t.Run("ReplacedStatPermissionDenied", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		replacePidFile(t, pidFile, strconv.Itoa(cmd.Process.Pid))

		defaultStatFd := statFd
		statFd = func(name string) (os.FileInfo, error) {
			return nil, &os.PathError{Op: "stat", Path: name, Err: syscall.EACCES}
		}
		t.Cleanup(func() { statFd = defaultStatFd })

		open, known := processHasFile(cmd.Process.Pid, pidFile, fileInfo(t, pidFile))
		assert.False(t, known)
		assert.False(t, open)
	})
	t.Run("ClosedMeanwhile", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")

		defaultReadFdLink := readFdLink
		readFdLink = func(name string) (string, error) {
			return "", &os.PathError{Op: "readlink", Path: name, Err: syscall.ENOENT}
		}
		t.Cleanup(func() { readFdLink = defaultReadFdLink })

		open, known := processHasFile(cmd.Process.Pid, pidFile, fileInfo(t, pidFile))
		assert.True(t, known)
		assert.True(t, open)
	})
}

// TestReplacedFileName covers the name under which /proc lists an open file that was replaced.
func TestReplacedFileName(t *testing.T) {
	t.Run("Absolute", func(t *testing.T) {
		dir := t.TempDir()
		assert.Equal(t, filepath.Join(dir, "photoprism.pid")+" (deleted)", replacedFileName(filepath.Join(dir, "photoprism.pid")))
	})
	t.Run("SymlinkedFolder", func(t *testing.T) {
		dir := t.TempDir()
		link := filepath.Join(t.TempDir(), "storage")
		require.NoError(t, os.Symlink(dir, link))
		assert.Equal(t, filepath.Join(dir, "photoprism.pid")+" (deleted)", replacedFileName(filepath.Join(link, "photoprism.pid")))
	})
	t.Run("Relative", func(t *testing.T) {
		wd, err := os.Getwd()
		require.NoError(t, err)
		wd, err = filepath.EvalSymlinks(wd)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(wd, "photoprism.pid")+" (deleted)", replacedFileName("photoprism.pid"))
	})
	t.Run("MissingFolder", func(t *testing.T) {
		assert.Equal(t, "", replacedFileName(filepath.Join(t.TempDir(), "missing", "photoprism.pid")))
	})
}

// TestFileReplaced covers detecting an open file that was replaced by another file.
func TestFileReplaced(t *testing.T) {
	t.Run("Removed", func(t *testing.T) {
		fileName := unlockedPidFile(t, 1234)
		f, err := os.Open(fileName) //nolint:gosec // Opens a test file.
		require.NoError(t, err)
		t.Cleanup(func() { _ = f.Close() })
		replacePidFile(t, fileName, "1234")
		open, err := f.Stat()
		require.NoError(t, err)
		assert.True(t, fileReplaced(open, fileInfo(t, fileName)))
	})
	t.Run("Linked", func(t *testing.T) {
		info := fileInfo(t, unlockedPidFile(t, 1234))
		assert.False(t, fileReplaced(info, info))
	})
	t.Run("OtherDevice", func(t *testing.T) {
		open := sysFileInfo{sys: &syscall.Stat_t{Dev: 1, Nlink: 0}}
		info := sysFileInfo{sys: &syscall.Stat_t{Dev: 2, Nlink: 1}}
		assert.False(t, fileReplaced(open, info))
	})
	t.Run("SameDevice", func(t *testing.T) {
		open := sysFileInfo{sys: &syscall.Stat_t{Dev: 2, Nlink: 0}}
		info := sysFileInfo{sys: &syscall.Stat_t{Dev: 2, Nlink: 1}}
		assert.True(t, fileReplaced(open, info))
	})
	t.Run("NoStat", func(t *testing.T) {
		info := sysFileInfo{sys: &syscall.Stat_t{Dev: 2, Nlink: 1}}
		assert.False(t, fileReplaced(sysFileInfo{}, info))
		assert.False(t, fileReplaced(sysFileInfo{sys: &syscall.Stat_t{Dev: 2}}, sysFileInfo{}))
	})
	t.Run("Nil", func(t *testing.T) {
		info := fileInfo(t, unlockedPidFile(t, 1234))
		assert.False(t, fileReplaced(nil, info))
		assert.False(t, fileReplaced(info, nil))
	})
}

// TestClosePidFileOnExec covers keeping the pid file the daemon inherited out of the programs it runs.
func TestClosePidFileOnExec(t *testing.T) {
	// subprocessHoldsPidFile runs the stand-in daemon in the mode and reports whether the program it runs holds the pid file.
	subprocessHoldsPidFile := func(t *testing.T, mode string) bool {
		pidFile, f := lockedPidFile(t)

		null, err := os.Open(os.DevNull)
		require.NoError(t, err)
		t.Cleanup(func() { _ = null.Close() })

		// Passes the pid file as fd 4, like go-daemon.
		_, stdout := startHelper(t, mode, nil, null, f)
		line, err := stdout.ReadString('\n')
		require.NoError(t, err)

		pid, err := strconv.Atoi(strings.TrimSpace(line))
		require.NoError(t, err, line)
		t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })

		waitForName(t, pid, "sleep")
		open, known := processHasFile(pid, pidFile, fileInfo(t, pidFile))
		require.True(t, known)

		return open
	}

	t.Run("Inherited", func(t *testing.T) {
		assert.True(t, subprocessHoldsPidFile(t, "subprocess"))
	})
	t.Run("CloseOnExec", func(t *testing.T) {
		assert.False(t, subprocessHoldsPidFile(t, "subprocess-cloexec"))
	})
}

// TestStartPrologue covers the pid file a daemon started by go-daemon inherits, and keeping it out of the programs it runs.
func TestStartPrologue(t *testing.T) {
	t.Run("Reborn", func(t *testing.T) {
		dir := t.TempDir()
		pidFile := filepath.Join(dir, "photoprism.pid")
		logFile := filepath.Join(dir, "photoprism.log")

		cmd, stdout := startHelper(t, "reborn", []string{stopHelperPidEnv + "=" + pidFile, stopHelperLogEnv + "=" + logFile})
		line, err := stdout.ReadString('\n')
		require.NoError(t, err)
		require.NoError(t, cmd.Wait())

		pid, err := strconv.Atoi(strings.TrimSpace(line))
		require.NoError(t, err, line)
		daemonProcess, err := os.FindProcess(pid)
		require.NoError(t, err)
		t.Cleanup(func() { _ = daemonProcess.Kill(); _ = daemonProcess.Release() })

		var report string

		require.Eventually(t, func() bool {
			b, readErr := os.ReadFile(logFile) //nolint:gosec // Reads the stand-in daemon log.
			report = string(b)
			return readErr == nil && strings.Contains(report, "pid-file=") && strings.HasSuffix(report, "\n")
		}, 10*time.Second, 20*time.Millisecond)

		assert.Contains(t, report, "pid-file=true cloexec-before=false cloexec-after=true err=<nil>")
		// The daemon wrote its own process ID, which the parent waited for.
		assertPidFileContent(t, pidFile, strconv.Itoa(pid))
	})
	t.Run("NotRebornFd", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		null, err := os.Open(os.DevNull)
		require.NoError(t, err)
		t.Cleanup(func() { _ = null.Close() })

		// Passes the file as fd 4, which a process that go-daemon did not start keeps as it is.
		cmd, stdout := startHelper(t, "prologue", nil, null, f)
		line, err := stdout.ReadString('\n')
		require.NoError(t, err)
		require.NoError(t, cmd.Wait())
		assert.Equal(t, "cloexec-before=false cloexec-after=false err=<nil>\n", line, pidFile)
		assertPidFileContent(t, pidFile, "")
	})
	t.Run("NotReborn", func(t *testing.T) {
		require.False(t, daemon.WasReborn())
		prev := event.SetAuditRecorder(func(event.Data) {})
		t.Cleanup(func() { event.SetAuditRecorder(prev) })
		require.NoError(t, startPrologue())
		assert.False(t, event.AuditRecordedSync())
	})
}

// TestPidFileHolders covers listing the processes that have a pid file open.
func TestPidFileHolders(t *testing.T) {
	t.Run("Holder", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		cmd := startProcess(t, "/opt/photoprism/bin/photoprism-test", "30", f)
		require.NoError(t, f.Close())
		holders, _ := pidFileHolders(fileInfo(t, pidFile))
		assert.Equal(t, []string{fmt.Sprintf("process %d (photoprism-test)", cmd.Process.Pid)}, holders)
		assert.True(t, strings.HasPrefix(pidFileHolderHint(pidFile, fileInfo(t, pidFile)), ", and is open in "+holders[0]))
	})
	t.Run("Self", func(t *testing.T) {
		requireProc(t)
		pidFile, _ := lockedPidFile(t)
		holders, _ := pidFileHolders(fileInfo(t, pidFile))
		assert.Empty(t, holders)
	})
	t.Run("ReplacedFile", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		replacePidFile(t, pidFile, "")
		holders, _ := pidFileHolders(fileInfo(t, pidFile))
		assert.NotContains(t, strings.Join(holders, ","), strconv.Itoa(cmd.Process.Pid))
	})
	t.Run("Unreadable", func(t *testing.T) {
		requireProc(t)
		defaultReadFdLink := readFdLink
		readFdLink = func(name string) (string, error) {
			return "", &os.PathError{Op: "readlink", Path: name, Err: syscall.EACCES}
		}
		t.Cleanup(func() { readFdLink = defaultReadFdLink })
		pidFile, f := lockedPidFile(t)
		startProcess(t, "photoprism-test", "30", f)
		holders, complete := pidFileHolders(fileInfo(t, pidFile))
		assert.Empty(t, holders)
		assert.False(t, complete)
		assert.Equal(t, "; run fuser -v "+pidFile+" or lsof "+pidFile+" to find the process that holds it", pidFileHolderHint(pidFile, fileInfo(t, pidFile)))
	})
	t.Run("PartlyReadable", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		cmd := startProcess(t, "photoprism-test", "30", f)
		other := startProcess(t, "sleep", "30")
		require.NoError(t, f.Close())

		// The open files of one other process cannot be read.
		defaultReadFdLink := readFdLink
		otherFd := filepath.Join("/proc", strconv.Itoa(other.Process.Pid), "fd")
		readFdLink = func(name string) (string, error) {
			if strings.HasPrefix(name, otherFd+"/") {
				return "", &os.PathError{Op: "readlink", Path: name, Err: syscall.EACCES}
			}
			return defaultReadFdLink(name)
		}
		t.Cleanup(func() { readFdLink = defaultReadFdLink })

		hint := pidFileHolderHint(pidFile, fileInfo(t, pidFile))
		assert.Contains(t, hint, fmt.Sprintf("is open in process %d (photoprism-test)", cmd.Process.Pid))
		assert.Contains(t, hint, "; run fuser -v "+pidFile)
	})
	t.Run("NameSanitized", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		cmd := startProcess(t, "photoprism-test\x1b[31m", "30", f)
		require.NoError(t, f.Close())
		hint := pidFileHolderHint(pidFile, fileInfo(t, pidFile))
		assert.Contains(t, hint, strconv.Itoa(cmd.Process.Pid))
		assert.NotContains(t, hint, "\x1b")
	})
}

// TestPidFileHolderHint covers naming the processes that have a pid file open within the time limit.
func TestPidFileHolderHint(t *testing.T) {
	t.Run("Timeout", func(t *testing.T) {
		defaultScan, defaultTimeout := scanHolders, holderScanTimeout
		release, done := make(chan struct{}), make(chan struct{})
		scanHolders = func(os.FileInfo) ([]string, bool) {
			defer close(done)
			<-release
			return []string{"process 1 (late)"}, true
		}
		holderScanTimeout = 100 * time.Millisecond
		t.Cleanup(func() {
			close(release)
			<-done
			scanHolders, holderScanTimeout = defaultScan, defaultTimeout
		})

		pidFile, _ := lockedPidFile(t)
		start := time.Now()
		assert.Equal(t, "; run fuser -v "+pidFile+" or lsof "+pidFile+" to find the process that holds it", pidFileHolderHint(pidFile, fileInfo(t, pidFile)))
		assert.Less(t, time.Since(start), 5*time.Second)
	})
}

// TestFuserHint covers suggesting fuser with a pid file name only where it is safe to paste in a shell.
func TestFuserHint(t *testing.T) {
	t.Run("SafePath", func(t *testing.T) {
		assert.Equal(t, "run fuser -v /photoprism/storage/photoprism-1.pid or lsof /photoprism/storage/photoprism-1.pid to find the process that holds it", fuserHint("/photoprism/storage/photoprism-1.pid"))
	})
	t.Run("UnsafePath", func(t *testing.T) {
		for _, pidFile := range []string{"/tmp/a b.pid", "/tmp/$(id).pid", "/tmp/a;b.pid", "/tmp/a'b.pid", "-rf.pid", "/tmp/ä.pid", "/tmp/a*b", ""} {
			assert.Equal(t, "run fuser -v or lsof on the pid file to find the process that holds it", fuserHint(pidFile), pidFile)
		}
	})
}

// TestUnconfirmedRemedy covers suggesting how to proceed where a process cannot be confirmed as the daemon.
func TestUnconfirmedRemedy(t *testing.T) {
	other := sysFileInfo{sys: &syscall.Stat_t{Uid: 4294967290}}

	t.Run("Owner", func(t *testing.T) {
		info := fileInfo(t, unlockedPidFile(t, 1234))
		assert.Equal(t, "; remove the pid file if no server is running", unconfirmedRemedy(info, false))
		assert.Equal(t, "; remove the pid file if no server is running", unconfirmedRemedy(info, true))
	})
	t.Run("OtherOwner", func(t *testing.T) {
		assert.Equal(t, "; run the command as uid 4294967290, who owns the pid file", unconfirmedRemedy(other, false))
	})
	t.Run("OtherOwnerMismatch", func(t *testing.T) {
		assert.Equal(t, "; run the command as uid 4294967290, who owns the pid file, or remove the pid file if no server is running",
			unconfirmedRemedy(other, true))
	})
	t.Run("NoOwner", func(t *testing.T) {
		assert.Equal(t, "", unconfirmedRemedy(sysFileInfo{}, true))
		assert.Equal(t, "", unconfirmedRemedy(nil, false))
	})
}

// TestFileOwnerName covers describing the owner of a file.
func TestFileOwnerName(t *testing.T) {
	t.Run("Current", func(t *testing.T) {
		current, err := user.Current()
		require.NoError(t, err)
		assert.Equal(t, "user "+current.Username, fileOwnerName(fileInfo(t, unlockedPidFile(t, 1234))))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, "uid 4294967290", fileOwnerName(sysFileInfo{sys: &syscall.Stat_t{Uid: 4294967290}}))
	})
	t.Run("NoStat", func(t *testing.T) {
		assert.Equal(t, "its owner", fileOwnerName(sysFileInfo{}))
		assert.Equal(t, "its owner", fileOwnerName(nil))
	})
}

// TestProcessOpenFile covers detecting an open file, optionally including the file it replaced.
func TestProcessOpenFile(t *testing.T) {
	t.Run("Current", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		open, known := processOpenFile(cmd.Process.Pid, fileInfo(t, pidFile), "")
		assert.True(t, known)
		assert.True(t, open)
	})
	t.Run("ReplacedNotMatched", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		replacePidFile(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		open, known := processOpenFile(cmd.Process.Pid, fileInfo(t, pidFile), "")
		assert.True(t, known)
		assert.False(t, open)
	})
	t.Run("ReplacedMatched", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		replacePidFile(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		open, known := processOpenFile(cmd.Process.Pid, fileInfo(t, pidFile), replacedFileName(pidFile))
		assert.True(t, known)
		assert.True(t, open)
	})
	t.Run("NoInfo", func(t *testing.T) {
		_, known := processOpenFile(os.Getpid(), nil, "")
		assert.False(t, known)
	})
}

// TestPidFileLocked covers detecting whether another process holds the lock on a pid file.
func TestPidFileLocked(t *testing.T) {
	t.Run("Held", func(t *testing.T) {
		pidFile, _ := lockedPidFile(t)
		held, known := pidFileLocked(pidFile, fileInfo(t, pidFile))
		assert.True(t, known)
		assert.True(t, held)
	})
	t.Run("Free", func(t *testing.T) {
		pidFile := unlockedPidFile(t, os.Getpid())
		held, known := pidFileLocked(pidFile, fileInfo(t, pidFile))
		assert.True(t, known)
		assert.False(t, held)
	})
	t.Run("Untestable", func(t *testing.T) {
		defaultProbe := probePidFileLock
		probePidFileLock = func(*os.File) error { return syscall.ENOLCK }
		t.Cleanup(func() { probePidFileLock = defaultProbe })
		pidFile := unlockedPidFile(t, os.Getpid())
		held, known := pidFileLocked(pidFile, fileInfo(t, pidFile))
		assert.False(t, known)
		assert.False(t, held)
	})
	t.Run("Shared", func(t *testing.T) {
		pidFile := unlockedPidFile(t, os.Getpid())
		sharedLock(t, pidFile)
		held, known := pidFileLocked(pidFile, fileInfo(t, pidFile))
		assert.True(t, known)
		assert.False(t, held)
	})
	t.Run("ReadOnly", func(t *testing.T) {
		pidFile, _ := lockedPidFile(t)
		require.NoError(t, os.Chmod(pidFile, 0o400))
		held, known := pidFileLocked(pidFile, fileInfo(t, pidFile))
		assert.True(t, known)
		assert.True(t, held)
	})
	t.Run("Missing", func(t *testing.T) {
		held, known := pidFileLocked(filepath.Join(t.TempDir(), "photoprism.pid"), nil)
		assert.False(t, known)
		assert.False(t, held)
	})
}

// TestClearStalePidFile covers emptying a pid file only while no other process holds its lock.
func TestClearStalePidFile(t *testing.T) {
	t.Run("Free", func(t *testing.T) {
		pidFile := unlockedPidFile(t, deadPid)
		info := keepFile(t, pidFile)
		require.NoError(t, clearStalePidFile(pidFile, info))
		assertEmptyPidFile(t, pidFile)
		assert.True(t, os.SameFile(info, fileInfo(t, pidFile)))
	})
	t.Run("Locked", func(t *testing.T) {
		pidFile, _ := lockedPidFile(t)
		require.NoError(t, fs.WriteString(pidFile, deadPidText))
		assert.ErrorIs(t, clearStalePidFile(pidFile, fileInfo(t, pidFile)), daemon.ErrWouldBlock)
		assertPidFileContent(t, pidFile, deadPidText)
	})
	t.Run("Missing", func(t *testing.T) {
		assert.NoError(t, clearStalePidFile(filepath.Join(t.TempDir(), "photoprism.pid"), nil))
	})
	t.Run("Symlink", func(t *testing.T) {
		target := unlockedPidFile(t, deadPid)
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.Symlink(target, pidFile))
		err := clearStalePidFile(pidFile, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be a symbolic link")
		assertPidFileContent(t, target, deadPidText)
	})
	t.Run("HardLink", func(t *testing.T) {
		pidFile := unlockedPidFile(t, deadPid)
		other := filepath.Join(t.TempDir(), "other.pid")
		require.NoError(t, os.Link(pidFile, other))
		assert.ErrorIs(t, clearStalePidFile(pidFile, fileInfo(t, pidFile)), errPidFileLinked)
		assertPidFileContent(t, pidFile, deadPidText)
		assertPidFileContent(t, other, deadPidText)
	})
	t.Run("LockUnavailable", func(t *testing.T) {
		pidFile := unlockedPidFile(t, deadPid)

		defaultLock := lockStalePidFile
		lockStalePidFile = func(*daemon.LockFile) error { return syscall.ENOLCK }
		t.Cleanup(func() { lockStalePidFile = defaultLock })

		require.NoError(t, clearStalePidFile(pidFile, fileInfo(t, pidFile)))
		assertEmptyPidFile(t, pidFile)
	})
	t.Run("Replaced", func(t *testing.T) {
		pidFile := unlockedPidFile(t, deadPid)
		info := fileInfo(t, pidFile)
		// Keeps the old file, so that the new one cannot reuse its inode.
		require.NoError(t, os.Rename(pidFile, pidFile+".old"))
		require.NoError(t, fs.WriteString(pidFile, deadPidText))
		err := clearStalePidFile(pidFile, info)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "was replaced")
		assertPidFileContent(t, pidFile, deadPidText)
		assertPidFileContent(t, pidFile+".old", deadPidText)
	})
	t.Run("ReplacedBeforeLock", func(t *testing.T) {
		pidFile := unlockedPidFile(t, deadPid)
		info := fileInfo(t, pidFile)

		defaultLock := lockStalePidFile
		lockStalePidFile = func(lock *daemon.LockFile) error {
			require.NoError(t, os.Rename(pidFile, pidFile+".old"))
			require.NoError(t, fs.WriteString(pidFile, deadPidText))
			return defaultLock(lock)
		}
		t.Cleanup(func() { lockStalePidFile = defaultLock })

		err := clearStalePidFile(pidFile, info)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "was replaced")
		assertPidFileContent(t, pidFile, deadPidText)
		assertPidFileContent(t, pidFile+".old", deadPidText)
	})
}

// TestClearExitedPidFile covers emptying the pid file of a daemon once its lock is released.
func TestClearExitedPidFile(t *testing.T) {
	t.Run("Released", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		require.NoError(t, fs.WriteString(pidFile, deadPidText))
		time.AfterFunc(200*time.Millisecond, func() { _ = f.Close() })
		require.NoError(t, clearExitedPidFile(pidFile, fileInfo(t, pidFile), 5*time.Second))
		assertEmptyPidFile(t, pidFile)
	})
	t.Run("Held", func(t *testing.T) {
		pidFile, _ := lockedPidFile(t)
		require.NoError(t, fs.WriteString(pidFile, deadPidText))
		start := time.Now()
		assert.ErrorIs(t, clearExitedPidFile(pidFile, fileInfo(t, pidFile), 300*time.Millisecond), daemon.ErrWouldBlock)
		assert.GreaterOrEqual(t, time.Since(start), 300*time.Millisecond)
		assertPidFileContent(t, pidFile, deadPidText)
	})
	t.Run("Replaced", func(t *testing.T) {
		pidFile := unlockedPidFile(t, deadPid)
		info := fileInfo(t, pidFile)
		require.NoError(t, os.Rename(pidFile, pidFile+".old"))
		require.NoError(t, fs.WriteString(pidFile, deadPidText))
		start := time.Now()
		require.Error(t, clearExitedPidFile(pidFile, info, 5*time.Second))
		assert.Less(t, time.Since(start), time.Second)
		assertPidFileContent(t, pidFile, deadPidText)
	})
}

// TestChildAlreadyRunning covers detecting a running daemon before starting another one.
func TestChildAlreadyRunning(t *testing.T) {
	t.Run("Running", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		pid, running, err := childAlreadyRunning(pidFile)
		require.NoError(t, err)
		assert.Equal(t, cmd.Process.Pid, pid)
		assert.True(t, running)
	})
	t.Run("Stale", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		_, running, err := childAlreadyRunning(unlockedPidFile(t, cmd.Process.Pid))
		require.NoError(t, err)
		assert.False(t, running)
	})
	t.Run("NoPidFile", func(t *testing.T) {
		_, running, err := childAlreadyRunning(filepath.Join(t.TempDir(), "photoprism.pid"))
		require.NoError(t, err)
		assert.False(t, running)
	})
	t.Run("Empty", func(t *testing.T) {
		for _, content := range []string{"", "\n"} {
			pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
			require.NoError(t, fs.WriteString(pidFile, content))
			pid, running, err := childAlreadyRunning(pidFile)
			require.NoError(t, err, content)
			assert.False(t, running)
			assert.Equal(t, 0, pid)
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, content := range []string{"not a pid", "\x00\x00\x00\x00", strings.Repeat("x", pidFileMaxSize)} {
			pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
			require.NoError(t, fs.WriteString(pidFile, content))
			_, running, err := childAlreadyRunning(pidFile)
			require.NoError(t, err, content)
			assert.False(t, running)
		}
	})
	t.Run("TooLarge", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		content := strings.Repeat("1", pidFileMaxSize+1)
		require.NoError(t, fs.WriteString(pidFile, content))
		_, running, err := childAlreadyRunning(pidFile)
		require.ErrorIs(t, err, errPidFileTooLarge)
		assert.False(t, running)
		assertPidFileContent(t, pidFile, content)
	})
	t.Run("InvalidLocked", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		_, err := f.WriteString("not a pid")
		require.NoError(t, err)
		_, running, err := childAlreadyRunning(pidFile)
		require.Error(t, err)
		assert.ErrorIs(t, err, errPidFileInvalid)
		assert.Contains(t, err.Error(), "another process holds it locked")
		assert.False(t, running)
	})
	t.Run("Directory", func(t *testing.T) {
		_, running, err := childAlreadyRunning(t.TempDir())
		require.Error(t, err)
		assert.False(t, running)
	})
	t.Run("Replaced", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		replacePidFile(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		pid, running, err := childAlreadyRunning(pidFile)
		require.NoError(t, err)
		assert.Equal(t, cmd.Process.Pid, pid)
		assert.True(t, running)
	})
	t.Run("ReplacedOtherProgram", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		cmd := startProcess(t, "sleep", "30", f)
		require.NoError(t, f.Close())
		replacePidFile(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		_, running, err := childAlreadyRunning(pidFile)
		require.NoError(t, err)
		assert.False(t, running)
	})
}

// TestWaitForExit covers waiting for a daemon to exit.
func TestWaitForExit(t *testing.T) {
	t.Run("Exits", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "0.3")
		start := time.Now()
		require.NoError(t, waitForExit(cmd.Process, 10*time.Second))
		assert.GreaterOrEqual(t, time.Since(start), 100*time.Millisecond)
	})
	t.Run("OtherName", func(t *testing.T) {
		cmd := startProcess(t, "sleep", "0.3")
		start := time.Now()
		require.NoError(t, waitForExit(cmd.Process, 10*time.Second))
		assert.GreaterOrEqual(t, time.Since(start), 100*time.Millisecond)
	})
	t.Run("Timeout", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		err := waitForExit(cmd.Process, 300*time.Millisecond)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has not exited")
	})
}

// TestStopDaemon covers stopping the daemon that holds a pid file.
func TestStopDaemon(t *testing.T) {
	t.Run("Stops", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		info := keepFile(t, pidFile)
		require.NoError(t, stopDaemon(pidFile, 10*time.Second))
		assert.False(t, processRunning(cmd.Process))
		assertEmptyPidFile(t, pidFile)
		assert.True(t, os.SameFile(info, fileInfo(t, pidFile)))
	})
	t.Run("StopsWhileExiting", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		// The exiting daemon still holds the lock for the first attempts.
		defaultLock := lockStalePidFile
		attempts := 0
		lockStalePidFile = func(lock *daemon.LockFile) error {
			if attempts++; attempts <= 3 {
				return daemon.ErrWouldBlock
			}
			return defaultLock(lock)
		}
		t.Cleanup(func() { lockStalePidFile = defaultLock })
		require.NoError(t, stopDaemon(pidFile, 10*time.Second))
		assert.False(t, processRunning(cmd.Process))
		assert.Equal(t, 4, attempts)
		assertEmptyPidFile(t, pidFile)
	})
	t.Run("StopsAndRestarted", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		// Another daemon locks the same file before stop empties it.
		defaultLock := lockStalePidFile
		var restart sync.Once
		lockStalePidFile = func(lock *daemon.LockFile) error {
			restart.Do(func() {
				next, err := daemon.OpenLockFile(pidFile, fs.ModeFile)
				require.NoError(t, err)
				t.Cleanup(func() { _ = next.Close() })
				// Waits until the stopped daemon has closed its files.
				require.Eventually(t, func() bool { return next.Lock() == nil }, 5*time.Second, 10*time.Millisecond)
				require.NoError(t, fs.WriteString(pidFile, deadPidText))
			})
			return defaultLock(lock)
		}
		t.Cleanup(func() { lockStalePidFile = defaultLock })
		start := time.Now()
		require.NoError(t, stopDaemon(pidFile, 10*time.Second))
		assert.Less(t, time.Since(start), 5*time.Second)
		assert.False(t, processRunning(cmd.Process))
		assertPidFileContent(t, pidFile, deadPidText)
	})
	t.Run("Replaced", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		replacePidFile(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		require.NoError(t, stopDaemon(pidFile, 10*time.Second))
		assert.False(t, processRunning(cmd.Process))
		assertEmptyPidFile(t, pidFile)
	})
	t.Run("ReplacedOtherProgram", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		cmd := startProcess(t, "sleep", "30", f)
		require.NoError(t, f.Close())
		replacePidFile(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		require.NoError(t, stopDaemon(pidFile, 10*time.Second))
		assert.True(t, processRunning(cmd.Process))
		assertEmptyPidFile(t, pidFile)
	})
	t.Run("StalePhotoPrism", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		pidFile := unlockedPidFile(t, cmd.Process.Pid)
		info := keepFile(t, pidFile)
		require.NoError(t, stopDaemon(pidFile, 10*time.Second))
		assertEmptyPidFile(t, pidFile)
		assert.True(t, os.SameFile(info, fileInfo(t, pidFile)))
		assert.True(t, processRunning(cmd.Process))
	})
	t.Run("LockedByOther", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		pidFile, _ := lockedPidFile(t)
		require.NoError(t, fs.WriteString(pidFile, strconv.Itoa(cmd.Process.Pid)))
		err := stopDaemon(pidFile, 10*time.Second)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "is not a PhotoPrism daemon holding it")
		assertPidFileContent(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		assert.True(t, processRunning(cmd.Process))
	})
	t.Run("Self", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		cmd, stdout := startHelper(t, "stop-self", []string{stopHelperPidEnv + "=" + pidFile})
		line, err := stdout.ReadString('\n')
		require.NoError(t, err)
		assert.Contains(t, line, "is not a PhotoPrism daemon holding it")
		require.NoError(t, cmd.Wait())
	})
	t.Run("HeldByOtherProgram", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		cmd := startProcess(t, "sleep", "30", f)
		require.NoError(t, fs.WriteString(pidFile, strconv.Itoa(cmd.Process.Pid)))
		require.Error(t, stopDaemon(pidFile, 10*time.Second))
		assertPidFileContent(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		assert.True(t, processRunning(cmd.Process))
	})
	t.Run("Symlink", func(t *testing.T) {
		cmd, target := startDaemon(t, "30")
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.Symlink(target, pidFile))
		err := stopDaemon(pidFile, 10*time.Second)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be a symbolic link")
		assert.True(t, processRunning(cmd.Process))
	})
	t.Run("DeadPid", func(t *testing.T) {
		pidFile := unlockedPidFile(t, deadPid)
		require.NoError(t, stopDaemon(pidFile, time.Second))
		assertEmptyPidFile(t, pidFile)
	})
	t.Run("Unconfirmed", func(t *testing.T) {
		defaultHoldsFile, defaultProbe := holdsFile, probePidFileLock
		holdsFile = func(int, string, os.FileInfo) (bool, bool) { return false, false }
		probePidFileLock = func(*os.File) error { return syscall.ENOLCK }
		t.Cleanup(func() { holdsFile, probePidFileLock = defaultHoldsFile, defaultProbe })

		cmd := startProcess(t, "photoprism-test", "30")
		pidFile := unlockedPidFile(t, cmd.Process.Pid)
		err := stopDaemon(pidFile, time.Second)
		require.ErrorIs(t, err, errDaemonUnconfirmed)
		assert.Contains(t, err.Error(), "cannot confirm that process")
		assertPidFileContent(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		assert.True(t, processRunning(cmd.Process))

		_, running, err := childAlreadyRunning(pidFile)
		assert.ErrorIs(t, err, errDaemonUnconfirmed)
		assert.False(t, running)
	})
	t.Run("FallbackKeepsFile", func(t *testing.T) {
		defaultHoldsFile := holdsFile
		holdsFile = func(int, string, os.FileInfo) (bool, bool) { return false, false }
		t.Cleanup(func() { holdsFile = defaultHoldsFile })

		// The stand-in matches the pid file, which nothing holds locked, e.g. because it was replaced.
		cmd := startProcess(t, "photoprism-test", "30")
		pidFile := unlockedPidFile(t, cmd.Process.Pid)
		err := stopDaemon(pidFile, time.Second)
		require.ErrorIs(t, err, errDaemonUnconfirmed)
		assert.Contains(t, err.Error(), "the pid file is not locked")
		// The test runs as the owner of the pid file.
		assert.True(t, strings.HasSuffix(err.Error(), "the pid file is not locked; remove the pid file if no server is running"), err.Error())
		assertPidFileContent(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		assert.True(t, processRunning(cmd.Process))
	})
	t.Run("FallbackOtherUser", func(t *testing.T) {
		defaultHoldsFile, defaultProcessUid := holdsFile, processUid
		holdsFile = func(int, string, os.FileInfo) (bool, bool) { return false, false }
		t.Cleanup(func() { holdsFile, processUid = defaultHoldsFile, defaultProcessUid })

		cmd := startProcess(t, "photoprism-test", "30")
		pidFile := unlockedPidFile(t, cmd.Process.Pid)
		otherUid := fileOwner(t, fileInfo(t, pidFile)) + 1
		processUid = func(int) (uint32, error) { return otherUid, nil }
		err := stopDaemon(pidFile, time.Second)
		require.ErrorIs(t, err, errDaemonUnconfirmed)
		assert.Contains(t, err.Error(), "it does not match the pid file; remove the pid file if no server is running")
		assertPidFileContent(t, pidFile, strconv.Itoa(cmd.Process.Pid))
	})
	t.Run("FallbackStops", func(t *testing.T) {
		defaultHoldsFile := holdsFile
		holdsFile = func(int, string, os.FileInfo) (bool, bool) { return false, false }
		t.Cleanup(func() { holdsFile = defaultHoldsFile })

		cmd, pidFile := startDaemon(t, "30")
		require.NoError(t, stopDaemon(pidFile, 10*time.Second))
		assert.False(t, processRunning(cmd.Process))
		// A daemon identified by its lock leaves the pid file as it is.
		assertPidFileContent(t, pidFile, strconv.Itoa(cmd.Process.Pid))
	})
	t.Run("StaleHardLink", func(t *testing.T) {
		pidFile := unlockedPidFile(t, deadPid)
		other := filepath.Join(t.TempDir(), "other.pid")
		require.NoError(t, os.Link(pidFile, other))
		require.NoError(t, stopDaemon(pidFile, time.Second))
		assertPidFileContent(t, pidFile, deadPidText)
		assertPidFileContent(t, other, deadPidText)
	})
	t.Run("NoPidFile", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, stopDaemon(pidFile, time.Second))
		assert.NoFileExists(t, pidFile)
	})
	t.Run("Empty", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, fs.WriteString(pidFile, ""))
		info := keepFile(t, pidFile)
		require.NoError(t, stopDaemon(pidFile, time.Second))
		assertEmptyPidFile(t, pidFile)
		assert.True(t, os.SameFile(info, fileInfo(t, pidFile)))
	})
	t.Run("EmptyLocked", func(t *testing.T) {
		// The test process holds the lock, which is not listed as another process.
		pidFile, _ := lockedPidFile(t)
		err := stopDaemon(pidFile, time.Second)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "contains no process ID")
		assert.Contains(t, err.Error(), "run fuser -v "+pidFile)
		assertEmptyPidFile(t, pidFile)
	})
	t.Run("EmptyLockedHolder", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		cmd := startProcess(t, "photoprism-test", "30", f)
		require.NoError(t, f.Close())
		err := stopDaemon(pidFile, time.Second)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "contains no process ID")
		assert.Contains(t, err.Error(), fmt.Sprintf("is open in process %d (photoprism-test)", cmd.Process.Pid))
		assertEmptyPidFile(t, pidFile)
		assert.True(t, processRunning(cmd.Process))
	})
	t.Run("Timeout", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		cmd, stdout := startHelper(t, "ignore-term", nil, f)
		line, err := stdout.ReadString('\n')
		require.NoError(t, err)
		require.Equal(t, "ready\n", line)
		require.NoError(t, fs.WriteString(pidFile, strconv.Itoa(cmd.Process.Pid)))
		err = stopDaemon(pidFile, 500*time.Millisecond)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has not exited")
		assert.True(t, processRunning(cmd.Process))
	})
}

// fileOwner returns the user ID that owns the file.
func fileOwner(t *testing.T, info os.FileInfo) uint32 {
	t.Helper()

	st, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)

	return st.Uid
}

// sysFileInfo describes a file with the given system information.
type sysFileInfo struct {
	os.FileInfo
	sys     any
	modTime time.Time
}

// Sys returns the system information.
func (i sysFileInfo) Sys() any { return i.sys }

// ModTime returns the modification time.
func (i sysFileInfo) ModTime() time.Time { return i.modTime }

// daemonHolds returns whether the process holds the pid file as the daemon does, and requires that it can be confirmed.
func daemonHolds(t *testing.T, p *os.Process, pidFile string, info os.FileInfo) bool {
	t.Helper()

	held, _, err := daemonHoldsPidFile(p, pidFile, info)
	require.NoError(t, err)

	return held
}

// assertUnconfirmed checks that it cannot be confirmed whether the process holds the pid file.
func assertUnconfirmed(t *testing.T, p *os.Process, pidFile string, info os.FileInfo) {
	t.Helper()

	held, viaLock, err := daemonHoldsPidFile(p, pidFile, info)
	assert.ErrorIs(t, err, errDaemonUnconfirmed)
	assert.True(t, viaLock)
	assert.False(t, held)
}

// TestDaemonCandidate covers checking that a process matches a pid file as the daemon would.
func TestDaemonCandidate(t *testing.T) {
	// withUid stubs the user ID that the process runs as, and makes it lead its session from the Unix epoch.
	withUid := func(t *testing.T, uid uint32) {
		defaultProcessUid, defaultProcessStart := processUid, processStart
		processUid = func(int) (uint32, error) { return uid, nil }
		processStart = func(pid int) (int, time.Time, error) { return pid, time.Unix(0, 0), nil }
		t.Cleanup(func() { processUid, processStart = defaultProcessUid, defaultProcessStart })
	}

	// withStart stubs the session ID and start time of the process.
	withStart := func(t *testing.T, session int, started time.Time) {
		defaultProcessStart := processStart
		processStart = func(int) (int, time.Time, error) { return session, started, nil }
		t.Cleanup(func() { processStart = defaultProcessStart })
	}

	written := time.Unix(1700000000, 0)

	t.Run("OwnerNotGroup", func(t *testing.T) {
		withUid(t, 5)
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Gid: 7}, modTime: written}
		assert.True(t, daemonCandidate(os.Getpid(), info))
	})
	t.Run("GroupNotOwner", func(t *testing.T) {
		withUid(t, 7)
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Gid: 7}, modTime: written}
		assert.False(t, daemonCandidate(os.Getpid(), info))
	})
	t.Run("GroupWritable", func(t *testing.T) {
		withUid(t, 5)
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Gid: 7, Mode: syscall.S_IFREG | 0o660}, modTime: written}
		assert.False(t, daemonCandidate(os.Getpid(), info))
	})
	t.Run("OthersWritable", func(t *testing.T) {
		withUid(t, 5)
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Gid: 7, Mode: syscall.S_IFREG | 0o602}, modTime: written}
		assert.False(t, daemonCandidate(os.Getpid(), info))
	})
	t.Run("OwnerOnlyWritable", func(t *testing.T) {
		withUid(t, 5)
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Gid: 7, Mode: syscall.S_IFREG | 0o644, Nlink: 1}, modTime: written}
		assert.True(t, daemonCandidate(os.Getpid(), info))
	})
	t.Run("HardLink", func(t *testing.T) {
		withUid(t, 5)
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Gid: 7, Mode: syscall.S_IFREG | 0o640, Nlink: 2}, modTime: written}
		assert.False(t, daemonCandidate(os.Getpid(), info))
	})
	t.Run("NotSessionLeader", func(t *testing.T) {
		withUid(t, 5)
		withStart(t, os.Getpid()+1, time.Unix(0, 0))
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Nlink: 1}, modTime: written}
		assert.False(t, daemonCandidate(os.Getpid(), info))
	})
	t.Run("StartedAfterWrite", func(t *testing.T) {
		withUid(t, 5)
		withStart(t, os.Getpid(), written.Add(startTimeTolerance+time.Millisecond))
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Nlink: 1}, modTime: written}
		assert.False(t, daemonCandidate(os.Getpid(), info))
	})
	t.Run("StartedWithinTolerance", func(t *testing.T) {
		withUid(t, 5)
		withStart(t, os.Getpid(), written.Add(startTimeTolerance))
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Nlink: 1}, modTime: written}
		assert.True(t, daemonCandidate(os.Getpid(), info))
	})
	t.Run("StartUnreadable", func(t *testing.T) {
		requireProc(t)
		withUid(t, 5)
		defaultProcessStart := processStart
		processStart = func(int) (int, time.Time, error) { return 0, time.Time{}, os.ErrPermission }
		t.Cleanup(func() { processStart = defaultProcessStart })
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Nlink: 1}, modTime: written}
		assert.False(t, daemonCandidate(os.Getpid(), info))
	})
	t.Run("NoStat", func(t *testing.T) {
		withUid(t, 7)
		assert.True(t, daemonCandidate(os.Getpid(), sysFileInfo{}))
	})
	t.Run("NoInfo", func(t *testing.T) {
		assert.False(t, daemonCandidate(os.Getpid(), nil))
	})
	t.Run("Daemon", func(t *testing.T) {
		cmd := startProcess(t, "sleep", "30")
		assert.True(t, daemonCandidate(cmd.Process.Pid, fileInfo(t, unlockedPidFile(t, cmd.Process.Pid))))
	})
	t.Run("WrittenBeforeStart", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 1234)
		past := time.Now().Add(-time.Hour)
		require.NoError(t, os.Chtimes(pidFile, past, past))
		cmd := startProcess(t, "sleep", "30")
		assert.False(t, daemonCandidate(cmd.Process.Pid, fileInfo(t, pidFile)))
	})
	t.Run("OtherUser", func(t *testing.T) {
		info := fileInfo(t, unlockedPidFile(t, 1234))
		withUid(t, fileOwner(t, info)+1)
		assert.False(t, daemonCandidate(os.Getpid(), info))
	})
	t.Run("NotRunning", func(t *testing.T) {
		requireProc(t)
		assert.False(t, daemonCandidate(deadPid, fileInfo(t, unlockedPidFile(t, 1234))))
	})
}

// TestProcStart covers reading the session ID and start time of a process from /proc.
func TestProcStart(t *testing.T) {
	t.Run("SessionLeader", func(t *testing.T) {
		before := time.Now().Add(-startTimeTolerance)
		cmd := startProcess(t, "sleep", "30")
		session, started, err := procStart(cmd.Process.Pid)
		require.NoError(t, err)
		assert.Equal(t, cmd.Process.Pid, session)
		assert.False(t, started.Before(before.Add(-time.Second)), started)
		assert.False(t, started.After(time.Now().Add(startTimeTolerance)), started)
	})
	t.Run("Self", func(t *testing.T) {
		requireProc(t)
		session, _, err := procStart(os.Getpid())
		require.NoError(t, err)
		sid, err := unix.Getsid(0)
		require.NoError(t, err)
		assert.Equal(t, sid, session)
	})
	t.Run("NotRunning", func(t *testing.T) {
		requireProc(t)
		_, _, err := procStart(deadPid)
		assert.Error(t, err)
	})
}

// TestParseProcStat covers reading the session ID and start time from a /proc/<pid>/stat file.
func TestParseProcStat(t *testing.T) {
	fields := "S 1 4242 4243 0 -1 4194560 511 0 0 0 0 0 0 0 20 0 1 0 53834208 18079744 1895"

	t.Run("Success", func(t *testing.T) {
		session, ticks, err := parseProcStat([]byte("4242 (photoprism) " + fields + "\n"))
		require.NoError(t, err)
		assert.Equal(t, 4243, session)
		assert.Equal(t, uint64(53834208), ticks)
	})
	t.Run("NameWithParentheses", func(t *testing.T) {
		session, ticks, err := parseProcStat([]byte("4242 (a) 1 2 3 (b) " + fields))
		require.NoError(t, err)
		assert.Equal(t, 4243, session)
		assert.Equal(t, uint64(53834208), ticks)
	})
	t.Run("NoName", func(t *testing.T) {
		_, _, err := parseProcStat([]byte("4242 photoprism " + fields))
		assert.Error(t, err)
	})
	t.Run("TooFewFields", func(t *testing.T) {
		_, _, err := parseProcStat([]byte("4242 (photoprism) S 1 4242 4243"))
		assert.Error(t, err)
	})
	t.Run("InvalidSession", func(t *testing.T) {
		_, _, err := parseProcStat([]byte("4242 (photoprism) S 1 4242 x 0 -1 4194560 511 0 0 0 0 0 0 0 20 0 1 0 53834208"))
		assert.Error(t, err)
	})
	t.Run("InvalidStart", func(t *testing.T) {
		_, _, err := parseProcStat([]byte("4242 (photoprism) S 1 4242 4243 0 -1 4194560 511 0 0 0 0 0 0 0 20 0 1 0 -5"))
		assert.Error(t, err)
	})
}

// TestParseBootTime covers reading the boot time from /proc/stat.
func TestParseBootTime(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		boot, err := parseBootTime([]byte("cpu  1 2 3\nintr 5\nbtime 1791528824\nprocesses 10\n"))
		require.NoError(t, err)
		assert.Equal(t, time.Unix(1791528824, 0), boot)
	})
	t.Run("Missing", func(t *testing.T) {
		_, err := parseBootTime([]byte("cpu  1 2 3\n"))
		assert.Error(t, err)
	})
	t.Run("Invalid", func(t *testing.T) {
		_, err := parseBootTime([]byte("btime abc\n"))
		assert.Error(t, err)
	})
	t.Run("System", func(t *testing.T) {
		requireProc(t)
		b, err := os.ReadFile("/proc/stat")
		require.NoError(t, err)
		boot, err := parseBootTime(b)
		require.NoError(t, err)
		assert.True(t, boot.Before(time.Now()))
	})
}

// TestTicksDuration covers converting clock ticks from /proc into a duration.
func TestTicksDuration(t *testing.T) {
	t.Run("Ticks", func(t *testing.T) {
		assert.Equal(t, 1230*time.Millisecond, ticksDuration(123))
		assert.Equal(t, time.Duration(0), ticksDuration(0))
	})
	t.Run("LongUptime", func(t *testing.T) {
		// Ten years of uptime, which overflows when multiplied in nanoseconds first.
		ticks := uint64(10 * 365 * 24 * 3600 * procClockTicks)
		assert.Equal(t, 10*365*24*time.Hour, ticksDuration(ticks))
	})
	t.Run("Overflow", func(t *testing.T) {
		assert.Equal(t, time.Duration(math.MaxInt64), ticksDuration(math.MaxUint64))
	})
}

// TestProcUid covers reading the effective user ID of a process from /proc.
func TestProcUid(t *testing.T) {
	t.Run("Self", func(t *testing.T) {
		requireProc(t)
		uid, err := procUid(os.Getpid())
		require.NoError(t, err)
		assert.Equal(t, os.Geteuid(), int(uid))
	})
	t.Run("Root", func(t *testing.T) {
		requireProc(t)
		uid, err := procUid(1)
		require.NoError(t, err)
		assert.Equal(t, uint32(0), uid)
	})
	t.Run("NotRunning", func(t *testing.T) {
		requireProc(t)
		_, err := procUid(deadPid)
		assert.Error(t, err)
	})
}

// TestStatusUid covers reading the effective user ID from a /proc status file.
func TestStatusUid(t *testing.T) {
	t.Run("Effective", func(t *testing.T) {
		uid, err := statusUid([]byte("Name:\tphotoprism\nUid:\t1000\t1001\t1002\t1003\nGid:\t1000\t1000\t1000\t1000\n"))
		require.NoError(t, err)
		assert.Equal(t, uint32(1001), uid)
	})
	t.Run("Missing", func(t *testing.T) {
		_, err := statusUid([]byte("Name:\tphotoprism\nGid:\t1000\t1000\t1000\t1000\n"))
		assert.Error(t, err)
	})
	t.Run("Invalid", func(t *testing.T) {
		_, err := statusUid([]byte("Uid:\tabc\tdef\n"))
		assert.Error(t, err)
	})
}

// TestProcCmdline covers reading the command line of a process from /proc.
func TestProcCmdline(t *testing.T) {
	t.Run("Running", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		cmdline, err := procCmdline(cmd.Process.Pid)
		require.NoError(t, err)
		assert.Equal(t, "photoprism-test\x0030\x00", string(cmdline))
	})
	t.Run("NotRunning", func(t *testing.T) {
		requireProc(t)
		_, err := procCmdline(deadPid)
		assert.Error(t, err)
	})
}

// TestDaemonHoldsPidFile covers identifying the daemon, also where its open files cannot be listed.
func TestDaemonHoldsPidFile(t *testing.T) {
	t.Run("OpenFile", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		assert.True(t, daemonHolds(t, cmd.Process, pidFile, fileInfo(t, pidFile)))
	})
	t.Run("OpenFileOtherProgram", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		cmd := startProcess(t, "sleep", "30", f)
		assert.False(t, daemonHolds(t, cmd.Process, pidFile, fileInfo(t, pidFile)))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, daemonHolds(t, nil, filepath.Join(t.TempDir(), "photoprism.pid"), nil))
	})
	t.Run("FallbackOtherUser", func(t *testing.T) {
		defaultHoldsFile, defaultProcessUid := holdsFile, processUid
		holdsFile = func(int, string, os.FileInfo) (bool, bool) { return false, false }
		lockedFile, _ := lockedPidFile(t)
		otherUid := fileOwner(t, fileInfo(t, lockedFile)) + 1
		processUid = func(int) (uint32, error) { return otherUid, nil }
		t.Cleanup(func() { holdsFile, processUid = defaultHoldsFile, defaultProcessUid })

		daemonCmd := startProcess(t, "photoprism-test", "30")
		assertUnconfirmed(t, daemonCmd.Process, lockedFile, fileInfo(t, lockedFile))
	})
	t.Run("FallbackLockUntestable", func(t *testing.T) {
		defaultHoldsFile := holdsFile
		holdsFile = func(int, string, os.FileInfo) (bool, bool) { return false, false }
		t.Cleanup(func() { holdsFile = defaultHoldsFile })

		daemonCmd := startProcess(t, "photoprism-test", "30")
		pidFile := unlockedPidFile(t, daemonCmd.Process.Pid)
		info := fileInfo(t, pidFile)
		// The lock of a file that was replaced meanwhile cannot be tested.
		replacePidFile(t, pidFile, strconv.Itoa(daemonCmd.Process.Pid))
		_, known := pidFileLocked(pidFile, info)
		require.False(t, known)
		assertUnconfirmed(t, daemonCmd.Process, pidFile, info)
	})
	t.Run("FallbackNotSessionLeader", func(t *testing.T) {
		defaultHoldsFile, defaultProcessStart := holdsFile, processStart
		holdsFile = func(int, string, os.FileInfo) (bool, bool) { return false, false }
		processStart = func(pid int) (int, time.Time, error) { return pid + 1, time.Unix(0, 0), nil }
		t.Cleanup(func() { holdsFile, processStart = defaultHoldsFile, defaultProcessStart })

		daemonCmd := startProcess(t, "photoprism-test", "30")
		lockedFile, _ := lockedPidFile(t)
		assertUnconfirmed(t, daemonCmd.Process, lockedFile, fileInfo(t, lockedFile))
	})
	t.Run("FallbackSharedLock", func(t *testing.T) {
		defaultHoldsFile := holdsFile
		holdsFile = func(int, string, os.FileInfo) (bool, bool) { return false, false }
		t.Cleanup(func() { holdsFile = defaultHoldsFile })

		daemonCmd := startProcess(t, "photoprism-test", "30")
		pidFile := unlockedPidFile(t, daemonCmd.Process.Pid)
		reader := sharedLock(t, pidFile)
		require.NotNil(t, reader)
		assertUnconfirmed(t, daemonCmd.Process, pidFile, fileInfo(t, pidFile))
	})
	t.Run("Fallback", func(t *testing.T) {
		defaultHoldsFile := holdsFile
		holdsFile = func(int, string, os.FileInfo) (bool, bool) { return false, false }
		t.Cleanup(func() { holdsFile = defaultHoldsFile })

		daemonCmd := startProcess(t, "photoprism-test", "30")
		otherCmd := startProcess(t, "sleep", "30")
		lockedFile, _ := lockedPidFile(t)
		freeFile := unlockedPidFile(t, daemonCmd.Process.Pid)

		held, viaLock, err := daemonHoldsPidFile(daemonCmd.Process, lockedFile, fileInfo(t, lockedFile))
		require.NoError(t, err)
		assert.True(t, held)
		assert.True(t, viaLock)
		assert.False(t, daemonHolds(t, otherCmd.Process, lockedFile, fileInfo(t, lockedFile)))
		assertUnconfirmed(t, daemonCmd.Process, freeFile, fileInfo(t, freeFile))
	})
}

// TestOpenPidFile covers opening a pid file without following a symbolic link or blocking.
func TestOpenPidFile(t *testing.T) {
	t.Run("Regular", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 1234)
		f, info, err := openPidFile(pidFile, os.O_RDONLY)
		require.NoError(t, err)
		t.Cleanup(func() { _ = f.Close() })
		assert.True(t, info.Mode().IsRegular())
		assert.True(t, os.SameFile(fileInfo(t, pidFile), info))
	})
	t.Run("Symlink", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.Symlink(unlockedPidFile(t, 1234), pidFile))
		_, _, err := openPidFile(pidFile, os.O_RDONLY)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be a symbolic link")
	})
	t.Run("Directory", func(t *testing.T) {
		_, _, err := openPidFile(t.TempDir(), os.O_RDONLY)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "is not a regular file")
	})
	t.Run("Missing", func(t *testing.T) {
		_, _, err := openPidFile(filepath.Join(t.TempDir(), "photoprism.pid"), os.O_RDONLY)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
	t.Run("WriteOnly", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 1234)
		f, _, err := openPidFile(pidFile, os.O_WRONLY)
		require.NoError(t, err)
		t.Cleanup(func() { _ = f.Close() })
		require.NoError(t, f.Truncate(0))
		assertEmptyPidFile(t, pidFile)
	})
	t.Run("WriteOnlySymlink", func(t *testing.T) {
		target := unlockedPidFile(t, 1234)
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.Symlink(target, pidFile))
		_, _, err := openPidFile(pidFile, os.O_WRONLY)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be a symbolic link")
	})
	t.Run("AccessModeOnly", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		_, _, err := openPidFile(pidFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC)
		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.NoFileExists(t, pidFile)
	})
	t.Run("AccessModeOnlyExisting", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 1234)
		f, _, err := openPidFile(pidFile, os.O_RDWR|os.O_TRUNC|os.O_APPEND)
		require.NoError(t, err)
		_, err = f.WriteString("5")
		require.NoError(t, err)
		require.NoError(t, f.Close())
		assertPidFileContent(t, pidFile, "5234")
	})
	t.Run("WriteOnlyMissing", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		_, _, err := openPidFile(pidFile, os.O_WRONLY)
		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.NoFileExists(t, pidFile)
	})
}

// TestOpenSamePidFile covers opening a pid file only if it is still the expected file.
func TestOpenSamePidFile(t *testing.T) {
	t.Run("Same", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 1234)
		f, err := openSamePidFile(pidFile, os.O_RDONLY, fileInfo(t, pidFile))
		require.NoError(t, err)
		_ = f.Close()
	})
	t.Run("Replaced", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 1234)
		info := fileInfo(t, pidFile)
		// Keeps the old file, so that the new one cannot reuse its inode.
		require.NoError(t, os.Rename(pidFile, pidFile+".old"))
		require.NoError(t, fs.WriteString(pidFile, "1234"))
		_, err := openSamePidFile(pidFile, os.O_RDONLY, info)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "was replaced")
	})
	t.Run("NoInfo", func(t *testing.T) {
		f, err := openSamePidFile(unlockedPidFile(t, 1234), os.O_RDONLY, nil)
		require.NoError(t, err)
		_ = f.Close()
	})
}

// TestProcessRunning covers detecting whether a process still runs.
func TestProcessRunning(t *testing.T) {
	t.Run("Running", func(t *testing.T) {
		cmd := startProcess(t, "sleep", "30")
		assert.True(t, processRunning(cmd.Process))
	})
	t.Run("Exited", func(t *testing.T) {
		cmd := startProcess(t, "sleep", "0.1")
		require.Eventually(t, func() bool { return !processRunning(cmd.Process) }, 5*time.Second, 10*time.Millisecond)
	})
	t.Run("Zombie", func(t *testing.T) {
		requireProc(t)

		// Starts a process that is not reaped until the test ends.
		cmd := exec.Command("sleep", "0") //nolint:gosec // Runs sleep from PATH with a fixed duration.
		require.NoError(t, cmd.Start())
		t.Cleanup(func() { _ = cmd.Wait() })

		stat := filepath.Join("/proc", strconv.Itoa(cmd.Process.Pid), "stat")
		require.Eventually(t, func() bool {
			b, err := os.ReadFile(stat) //nolint:gosec // Reads the stand-in process status.
			_, state, _ := strings.Cut(string(b), ") ")
			return err == nil && strings.HasPrefix(state, "Z")
		}, 5*time.Second, 10*time.Millisecond)

		p, err := os.FindProcess(cmd.Process.Pid)
		require.NoError(t, err)
		t.Cleanup(func() { _ = p.Release() })
		require.NoError(t, p.Signal(syscall.Signal(0)))
		assert.False(t, processRunning(p))
	})
	t.Run("Starting", func(t *testing.T) {
		requireProc(t)

		// The command line of a process that is still being executed can be empty.
		for i := 0; i < 20; i++ {
			cmd := exec.Command("sleep", "30") //nolint:gosec // Runs sleep from PATH with a fixed duration.
			require.NoError(t, cmd.Start())
			running := processRunning(cmd.Process)
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			require.True(t, running, "attempt %d", i)
		}
	})
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, processRunning(nil))
	})
}

// TestProcStatFields covers splitting the fields after the program name from a /proc/<pid>/stat file.
func TestProcStatFields(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		rest, err := procStatFields([]byte("123 (a) b (c)) S 1"))
		require.NoError(t, err)
		assert.Equal(t, " S 1", string(rest))
	})
	t.Run("NoProgramName", func(t *testing.T) {
		_, err := procStatFields([]byte("123 photoprism S 1"))
		assert.Error(t, err)
	})
}

// TestParseProcState covers reading the process state from a /proc/<pid>/stat file.
func TestParseProcState(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		state, err := parseProcState([]byte("123 (photoprism) S 1 123 123 0 -1"))
		require.NoError(t, err)
		assert.Equal(t, byte('S'), state)
	})
	t.Run("NameWithParentheses", func(t *testing.T) {
		state, err := parseProcState([]byte("123 (a) Z (b)) S 1 123 123 0 -1"))
		require.NoError(t, err)
		assert.Equal(t, byte('S'), state)
	})
	t.Run("NameWithSpaces", func(t *testing.T) {
		state, err := parseProcState([]byte("123 (Web Content) R 1 123 123 0 -1"))
		require.NoError(t, err)
		assert.Equal(t, byte('R'), state)
	})
	t.Run("InvalidStat", func(t *testing.T) {
		for _, stat := range []string{"", "123 photoprism S", "123 (photoprism)", "123 (photoprism) ", "123 (photoprism)S 1"} {
			_, err := parseProcState([]byte(stat))
			assert.Error(t, err, "stat %q", stat)
		}
	})
}

// TestProcState covers reading the state of a process from /proc.
func TestProcState(t *testing.T) {
	t.Run("Stopped", func(t *testing.T) {
		cmd := startProcess(t, "sleep", "30")
		require.NoError(t, cmd.Process.Signal(syscall.SIGSTOP))
		t.Cleanup(func() { _ = cmd.Process.Signal(syscall.SIGCONT) })

		require.Eventually(t, func() bool {
			state, err := procState(cmd.Process.Pid)
			return err == nil && state == 'T'
		}, 5*time.Second, 10*time.Millisecond)
	})
	t.Run("Running", func(t *testing.T) {
		cmd := startProcess(t, "sleep", "30")
		state, err := procState(cmd.Process.Pid)
		require.NoError(t, err)
		assert.Contains(t, []byte{'R', 'S'}, state)
	})
	t.Run("NotFound", func(t *testing.T) {
		requireProc(t)
		_, err := procState(math.MaxInt32)
		assert.Error(t, err)
	})
}
