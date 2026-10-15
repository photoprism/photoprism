package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
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

// stopHelperEnv selects the stand-in process the test binary runs as for the stop tests.
const stopHelperEnv = "PHOTOPRISM_TEST_STOP_HELPER"

// stopHelperPidEnv names the pid file of the stop-self stand-in.
const stopHelperPidEnv = "PHOTOPRISM_TEST_STOP_PID_FILE"

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
	requireProc(t)

	cmd := exec.Command(os.Args[0], "-test.run=^$") //nolint:gosec // Runs this test binary as a stand-in.
	cmd.Args[0] = "photoprism-test"
	cmd.Env = append(append(os.Environ(), stopHelperEnv+"="+mode), env...)
	cmd.ExtraFiles = files
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
			_, _, err := readPidFile(pidFile)
			assert.Error(t, err, content)
			assert.NotErrorIs(t, err, errPidFileEmpty, content)
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
		_, _, err := readPidFile(pidFile)
		require.Error(t, err)
		assert.NotErrorIs(t, err, errPidFileEmpty)
		assert.Contains(t, err.Error(), "does not contain a valid process ID")
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
		p, err := os.FindProcess(3999999)
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
		pidFile := unlockedPidFile(t, 3999999)
		open, known := processHasFile(3999999, pidFile, fileInfo(t, pidFile))
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
		assert.True(t, daemonHoldsPidFile(cmd.Process, pidFile, fileInfo(t, pidFile)))
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
	t.Run("Missing", func(t *testing.T) {
		held, known := pidFileLocked(filepath.Join(t.TempDir(), "photoprism.pid"), nil)
		assert.False(t, known)
		assert.False(t, held)
	})
}

// TestClearStalePidFile covers emptying a pid file only while no other process holds its lock.
func TestClearStalePidFile(t *testing.T) {
	t.Run("Free", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 3999999)
		info := keepFile(t, pidFile)
		require.NoError(t, clearStalePidFile(pidFile, info))
		assertEmptyPidFile(t, pidFile)
		assert.True(t, os.SameFile(info, fileInfo(t, pidFile)))
	})
	t.Run("Locked", func(t *testing.T) {
		pidFile, _ := lockedPidFile(t)
		require.NoError(t, fs.WriteString(pidFile, "3999999"))
		assert.ErrorIs(t, clearStalePidFile(pidFile, fileInfo(t, pidFile)), daemon.ErrWouldBlock)
		assertPidFileContent(t, pidFile, "3999999")
	})
	t.Run("Missing", func(t *testing.T) {
		assert.NoError(t, clearStalePidFile(filepath.Join(t.TempDir(), "photoprism.pid"), nil))
	})
	t.Run("Symlink", func(t *testing.T) {
		target := unlockedPidFile(t, 3999999)
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.Symlink(target, pidFile))
		err := clearStalePidFile(pidFile, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be a symbolic link")
		assertPidFileContent(t, target, "3999999")
	})
	t.Run("HardLink", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 3999999)
		other := filepath.Join(t.TempDir(), "other.pid")
		require.NoError(t, os.Link(pidFile, other))
		assert.ErrorIs(t, clearStalePidFile(pidFile, fileInfo(t, pidFile)), errPidFileLinked)
		assertPidFileContent(t, pidFile, "3999999")
		assertPidFileContent(t, other, "3999999")
	})
	t.Run("LockUnavailable", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 3999999)

		defaultLock := lockStalePidFile
		lockStalePidFile = func(*daemon.LockFile) error { return syscall.ENOLCK }
		t.Cleanup(func() { lockStalePidFile = defaultLock })

		require.NoError(t, clearStalePidFile(pidFile, fileInfo(t, pidFile)))
		assertEmptyPidFile(t, pidFile)
	})
	t.Run("Replaced", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 3999999)
		info := fileInfo(t, pidFile)
		// Keeps the old file, so that the new one cannot reuse its inode.
		require.NoError(t, os.Rename(pidFile, pidFile+".old"))
		require.NoError(t, fs.WriteString(pidFile, "3999999"))
		err := clearStalePidFile(pidFile, info)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "was replaced")
		assertPidFileContent(t, pidFile, "3999999")
		assertPidFileContent(t, pidFile+".old", "3999999")
	})
	t.Run("ReplacedBeforeLock", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 3999999)
		info := fileInfo(t, pidFile)

		defaultLock := lockStalePidFile
		lockStalePidFile = func(lock *daemon.LockFile) error {
			require.NoError(t, os.Rename(pidFile, pidFile+".old"))
			require.NoError(t, fs.WriteString(pidFile, "3999999"))
			return defaultLock(lock)
		}
		t.Cleanup(func() { lockStalePidFile = defaultLock })

		err := clearStalePidFile(pidFile, info)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "was replaced")
		assertPidFileContent(t, pidFile, "3999999")
		assertPidFileContent(t, pidFile+".old", "3999999")
	})
}

// TestClearExitedPidFile covers emptying the pid file of a daemon once its lock is released.
func TestClearExitedPidFile(t *testing.T) {
	t.Run("Released", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		require.NoError(t, fs.WriteString(pidFile, "3999999"))
		time.AfterFunc(200*time.Millisecond, func() { _ = f.Close() })
		require.NoError(t, clearExitedPidFile(pidFile, fileInfo(t, pidFile), 5*time.Second))
		assertEmptyPidFile(t, pidFile)
	})
	t.Run("Held", func(t *testing.T) {
		pidFile, _ := lockedPidFile(t)
		require.NoError(t, fs.WriteString(pidFile, "3999999"))
		start := time.Now()
		assert.ErrorIs(t, clearExitedPidFile(pidFile, fileInfo(t, pidFile), 300*time.Millisecond), daemon.ErrWouldBlock)
		assert.GreaterOrEqual(t, time.Since(start), 300*time.Millisecond)
		assertPidFileContent(t, pidFile, "3999999")
	})
	t.Run("Replaced", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 3999999)
		info := fileInfo(t, pidFile)
		require.NoError(t, os.Rename(pidFile, pidFile+".old"))
		require.NoError(t, fs.WriteString(pidFile, "3999999"))
		start := time.Now()
		require.Error(t, clearExitedPidFile(pidFile, info, 5*time.Second))
		assert.Less(t, time.Since(start), time.Second)
		assertPidFileContent(t, pidFile, "3999999")
	})
}

// TestChildAlreadyRunning covers detecting a running daemon before starting another one.
func TestChildAlreadyRunning(t *testing.T) {
	t.Run("Running", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		pid, running := childAlreadyRunning(pidFile)
		assert.Equal(t, cmd.Process.Pid, pid)
		assert.True(t, running)
	})
	t.Run("Stale", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		_, running := childAlreadyRunning(unlockedPidFile(t, cmd.Process.Pid))
		assert.False(t, running)
	})
	t.Run("NoPidFile", func(t *testing.T) {
		_, running := childAlreadyRunning(filepath.Join(t.TempDir(), "photoprism.pid"))
		assert.False(t, running)
	})
	t.Run("Empty", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, fs.WriteString(pidFile, ""))
		_, running := childAlreadyRunning(pidFile)
		assert.False(t, running)
	})
	t.Run("Replaced", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		replacePidFile(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		pid, running := childAlreadyRunning(pidFile)
		assert.Equal(t, cmd.Process.Pid, pid)
		assert.True(t, running)
	})
	t.Run("ReplacedOtherProgram", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		cmd := startProcess(t, "sleep", "30", f)
		require.NoError(t, f.Close())
		replacePidFile(t, pidFile, strconv.Itoa(cmd.Process.Pid))
		_, running := childAlreadyRunning(pidFile)
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
				require.NoError(t, fs.WriteString(pidFile, "3999999"))
			})
			return defaultLock(lock)
		}
		t.Cleanup(func() { lockStalePidFile = defaultLock })
		start := time.Now()
		require.NoError(t, stopDaemon(pidFile, 10*time.Second))
		assert.Less(t, time.Since(start), 5*time.Second)
		assert.False(t, processRunning(cmd.Process))
		assertPidFileContent(t, pidFile, "3999999")
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
		pidFile := unlockedPidFile(t, 3999999)
		require.NoError(t, stopDaemon(pidFile, time.Second))
		assertEmptyPidFile(t, pidFile)
	})
	t.Run("StaleHardLink", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 3999999)
		other := filepath.Join(t.TempDir(), "other.pid")
		require.NoError(t, os.Link(pidFile, other))
		require.NoError(t, stopDaemon(pidFile, time.Second))
		assertPidFileContent(t, pidFile, "3999999")
		assertPidFileContent(t, other, "3999999")
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
		pidFile, _ := lockedPidFile(t)
		err := stopDaemon(pidFile, time.Second)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "contains no process ID")
		assertEmptyPidFile(t, pidFile)
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
	sys any
}

// Sys returns the system information.
func (i sysFileInfo) Sys() any { return i.sys }

// TestProcessOwnsFile covers checking that a process runs as the owner of a file.
func TestProcessOwnsFile(t *testing.T) {
	// withUid stubs the user ID that the process runs as.
	withUid := func(t *testing.T, uid uint32) {
		defaultProcessUid := processUid
		processUid = func(int) (uint32, error) { return uid, nil }
		t.Cleanup(func() { processUid = defaultProcessUid })
	}

	t.Run("OwnerNotGroup", func(t *testing.T) {
		withUid(t, 5)
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Gid: 7}}
		assert.True(t, processOwnsFile(os.Getpid(), info))
	})
	t.Run("GroupNotOwner", func(t *testing.T) {
		withUid(t, 7)
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Gid: 7}}
		assert.False(t, processOwnsFile(os.Getpid(), info))
	})
	t.Run("GroupWritable", func(t *testing.T) {
		withUid(t, 5)
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Gid: 7, Mode: syscall.S_IFREG | 0o660}}
		assert.False(t, processOwnsFile(os.Getpid(), info))
	})
	t.Run("OthersWritable", func(t *testing.T) {
		withUid(t, 5)
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Gid: 7, Mode: syscall.S_IFREG | 0o602}}
		assert.False(t, processOwnsFile(os.Getpid(), info))
	})
	t.Run("OwnerOnlyWritable", func(t *testing.T) {
		withUid(t, 5)
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Gid: 7, Mode: syscall.S_IFREG | 0o644, Nlink: 1}}
		assert.True(t, processOwnsFile(os.Getpid(), info))
	})
	t.Run("HardLink", func(t *testing.T) {
		withUid(t, 5)
		info := sysFileInfo{sys: &syscall.Stat_t{Uid: 5, Gid: 7, Mode: syscall.S_IFREG | 0o640, Nlink: 2}}
		assert.False(t, processOwnsFile(os.Getpid(), info))
	})
	t.Run("NoStat", func(t *testing.T) {
		withUid(t, 7)
		assert.True(t, processOwnsFile(os.Getpid(), sysFileInfo{}))
	})
	t.Run("NoInfo", func(t *testing.T) {
		assert.False(t, processOwnsFile(os.Getpid(), nil))
	})
	t.Run("Owner", func(t *testing.T) {
		cmd := startProcess(t, "sleep", "30")
		assert.True(t, processOwnsFile(cmd.Process.Pid, fileInfo(t, unlockedPidFile(t, 1234))))
	})
	t.Run("OtherUser", func(t *testing.T) {
		info := fileInfo(t, unlockedPidFile(t, 1234))
		withUid(t, fileOwner(t, info)+1)
		assert.False(t, processOwnsFile(os.Getpid(), info))
	})
	t.Run("NotRunning", func(t *testing.T) {
		requireProc(t)
		assert.False(t, processOwnsFile(3999999, fileInfo(t, unlockedPidFile(t, 1234))))
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
		_, err := procUid(3999999)
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
		_, err := procCmdline(3999999)
		assert.Error(t, err)
	})
}

// TestDaemonHoldsPidFile covers identifying the daemon, also where its open files cannot be listed.
func TestDaemonHoldsPidFile(t *testing.T) {
	t.Run("OpenFile", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		assert.True(t, daemonHoldsPidFile(cmd.Process, pidFile, fileInfo(t, pidFile)))
	})
	t.Run("OpenFileOtherProgram", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		cmd := startProcess(t, "sleep", "30", f)
		assert.False(t, daemonHoldsPidFile(cmd.Process, pidFile, fileInfo(t, pidFile)))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, daemonHoldsPidFile(nil, filepath.Join(t.TempDir(), "photoprism.pid"), nil))
	})
	t.Run("FallbackOtherUser", func(t *testing.T) {
		defaultHoldsFile, defaultProcessUid := holdsFile, processUid
		holdsFile = func(int, string, os.FileInfo) (bool, bool) { return false, false }
		lockedFile, _ := lockedPidFile(t)
		otherUid := fileOwner(t, fileInfo(t, lockedFile)) + 1
		processUid = func(int) (uint32, error) { return otherUid, nil }
		t.Cleanup(func() { holdsFile, processUid = defaultHoldsFile, defaultProcessUid })

		daemonCmd := startProcess(t, "photoprism-test", "30")
		assert.False(t, daemonHoldsPidFile(daemonCmd.Process, lockedFile, fileInfo(t, lockedFile)))
	})
	t.Run("Fallback", func(t *testing.T) {
		defaultHoldsFile := holdsFile
		holdsFile = func(int, string, os.FileInfo) (bool, bool) { return false, false }
		t.Cleanup(func() { holdsFile = defaultHoldsFile })

		daemonCmd := startProcess(t, "photoprism-test", "30")
		otherCmd := startProcess(t, "sleep", "30")
		lockedFile, _ := lockedPidFile(t)
		freeFile := unlockedPidFile(t, daemonCmd.Process.Pid)

		assert.True(t, daemonHoldsPidFile(daemonCmd.Process, lockedFile, fileInfo(t, lockedFile)))
		assert.False(t, daemonHoldsPidFile(otherCmd.Process, lockedFile, fileInfo(t, lockedFile)))
		assert.False(t, daemonHoldsPidFile(daemonCmd.Process, freeFile, fileInfo(t, freeFile)))
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
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, processRunning(nil))
	})
}
