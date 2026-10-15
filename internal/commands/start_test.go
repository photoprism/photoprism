package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/sevlyar/go-daemon"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
)

func TestStartCommand(t *testing.T) {
	t.Run("AuditRecordedThroughHub", func(t *testing.T) {
		requireTestDb(t)
		prev := event.SetAuditRecorder(func(event.Data) {})
		t.Cleanup(func() { event.SetAuditRecorder(prev) })

		output, err := RunWithTestContext(StartCommand, []string{"start", "--config"})
		require.NoError(t, err)
		assert.Contains(t, output, "http-port")
		assert.False(t, event.AuditRecordedSync())
	})
	t.Run("StartReport", func(t *testing.T) {
		// The config tells the cluster bootstrap that this flag only shows the report.
		for _, args := range [][]string{{"--config"}, {"-c"}} {
			ctx := newCommandContext(t, StartCommand, args...)
			ctx.Command = StartCommand
			assert.True(t, config.NewConfig(ctx).StartReport(), args)
		}

		ctx := newCommandContext(t, StartCommand)
		ctx.Command = StartCommand
		assert.False(t, config.NewConfig(ctx).StartReport())
	})
}

func TestWarnNonASCIIUsernames(t *testing.T) {
	requireTestDb(t)

	orig := event.SystemLog
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	event.SystemLog = logger
	t.Cleanup(func() { event.SystemLog = orig })

	base := entity.CountNonASCIIUsernames()

	if base == 0 {
		assert.Equal(t, 0, warnNonASCIIUsernames())
		assert.Nil(t, hook.LastEntry())
	}

	newExactTestUser(t, "warn-josé", false)

	assert.Equal(t, base+1, warnNonASCIIUsernames())
	require.NotNil(t, hook.LastEntry())
	assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
	assert.Equal(t, "users: "+english.Plural(base+1, "account", "accounts")+" with characters other than ASCII in the username › rename with photoprism users mod --username", hook.LastEntry().Message)
}

// TestDaemonRunning covers checking the pid file before the start command starts a daemon.
func TestDaemonRunning(t *testing.T) {
	t.Run("Running", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		pid, running, err := daemonRunning(pidFile)
		require.NoError(t, err)
		assert.True(t, running)
		assert.Equal(t, cmd.Process.Pid, pid)
	})
	t.Run("NoPidFile", func(t *testing.T) {
		_, running, err := daemonRunning(filepath.Join(t.TempDir(), "photoprism.pid"))
		require.NoError(t, err)
		assert.False(t, running)
	})
	t.Run("Symlink", func(t *testing.T) {
		_, target := startDaemon(t, "30")
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.Symlink(target, pidFile))
		_, running, err := daemonRunning(pidFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be a symbolic link")
		assert.False(t, running)
	})
	t.Run("InvalidLocked", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		_, err := f.WriteString("not a pid")
		require.NoError(t, err)
		_, running, err := daemonRunning(pidFile)
		require.Error(t, err)
		assert.False(t, running)
	})
	t.Run("TooLarge", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.WriteFile(pidFile, []byte(strings.Repeat("1", pidFileMaxSize+1)), 0o600))
		_, running, err := daemonRunning(pidFile)
		require.ErrorIs(t, err, errPidFileTooLarge)
		assert.False(t, running)
	})
	t.Run("DanglingSymlink", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.Symlink(filepath.Join(t.TempDir(), "missing.pid"), pidFile))
		_, _, err := daemonRunning(pidFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be a symbolic link")
	})
}

// TestWriteDaemonPid covers the daemon writing its process ID through the pid file it inherited as fd 4.
func TestWriteDaemonPid(t *testing.T) {
	// writePid runs the prologue as a daemon with the file as fd 4 and the context naming pidFile on its standard
	// input, and returns its process ID and error.
	writePid := func(t *testing.T, f *os.File, pidFile string) (int, string) {
		return writePidMode(t, "write-pid", f, pidFile)
	}

	t.Run("Overwrites", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		_, err := f.WriteString("123456789012")
		require.NoError(t, err)
		pid, result := writePid(t, f, pidFile)
		assert.Equal(t, "<nil>", result)
		assertPidFileContent(t, pidFile, strconv.Itoa(pid))
	})
	t.Run("OtherFile", func(t *testing.T) {
		held := unlockedPidFile(t, 1234)
		f, err := os.OpenFile(held, os.O_RDWR, 0) //nolint:gosec // Opens a test file.
		require.NoError(t, err)
		t.Cleanup(func() { _ = f.Close() })
		_, result := writePid(t, f, unlockedPidFile(t, 5678))
		assert.Contains(t, result, "is not the file the daemon inherited")
		assertPidFileContent(t, held, "1234")
	})
	t.Run("Symlink", func(t *testing.T) {
		target, f := lockedPidFile(t)
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.Symlink(target, pidFile))
		_, result := writePid(t, f, pidFile)
		assert.Contains(t, result, "must not be a symbolic link")
		assertPidFileContent(t, target, "")
	})
	t.Run("NotRegular", func(t *testing.T) {
		null, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
		require.NoError(t, err)
		t.Cleanup(func() { _ = null.Close() })
		_, result := writePid(t, null, os.DevNull)
		assert.Contains(t, result, "not a regular file")
	})
	t.Run("HardLink", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		other := filepath.Join(t.TempDir(), "other.pid")
		require.NoError(t, os.WriteFile(pidFile, []byte("1234"), 0o600))
		require.NoError(t, os.Link(pidFile, other))
		_, result := writePid(t, f, pidFile)
		assert.Contains(t, result, "more than one link")
		assertPidFileContent(t, other, "1234")
	})
	t.Run("ReadOnly", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 1234)
		f, err := os.Open(pidFile) //nolint:gosec // Opens a test file.
		require.NoError(t, err)
		t.Cleanup(func() { _ = f.Close() })
		_, result := writePid(t, f, pidFile)
		assert.Contains(t, result, "failed to write process id")
		assertPidFileContent(t, pidFile, "1234")
	})
	t.Run("NoContext", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		null, err := os.Open(os.DevNull)
		require.NoError(t, err)
		t.Cleanup(func() { _ = null.Close() })
		cmd, stdout := startHelperInput(t, "write-pid", []string{"_GO_DAEMON=1"}, null, null, f)
		line, err := stdout.ReadString('\n')
		require.NoError(t, err)
		require.NoError(t, cmd.Wait())
		assert.Contains(t, line, "standard input is not a pipe")
		assertPidFileContent(t, pidFile, "")
	})
	t.Run("ShortWrite", func(t *testing.T) {
		pidFile, f := lockedPidFile(t)
		_, err := f.WriteString("123456789012")
		require.NoError(t, err)
		_, result := writePidMode(t, "write-pid-short", f, pidFile)
		assert.Contains(t, result, "short write")
		assertPidFileContent(t, pidFile, "")
	})
}

// writePidMode runs the prologue in the helper mode as a daemon with the file as fd 4 and the context naming pidFile
// on its standard input, and returns its process ID and error.
func writePidMode(t *testing.T, mode string, f *os.File, pidFile string) (int, string) {
	t.Helper()

	null, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() { _ = null.Close() })

	context, err := json.Marshal(daemonContext{PidFileName: pidFile})
	require.NoError(t, err)

	cmd, stdout := startHelperInput(t, mode, []string{"_GO_DAEMON=1"}, bytes.NewReader(context), null, f)
	line, err := stdout.ReadString('\n')
	require.NoError(t, err)
	require.NoError(t, cmd.Wait())

	pidText, result, found := strings.Cut(strings.TrimPrefix(strings.TrimSpace(line), "pid="), " err=")
	require.True(t, found, line)
	pid, err := strconv.Atoi(pidText)
	require.NoError(t, err, line)

	return pid, result
}

// TestDaemonPidFileName covers reading the pid file name from the context go-daemon passes on standard input.
func TestDaemonPidFileName(t *testing.T) {
	// pipe returns the read end of a pipe that delivers the content.
	pipe := func(t *testing.T, content string) *os.File {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		t.Cleanup(func() { _ = r.Close() })
		_, err = w.WriteString(content)
		require.NoError(t, err)
		require.NoError(t, w.Close())
		return r
	}

	t.Run("Success", func(t *testing.T) {
		name, err := daemonPidFileName(pipe(t, `{"PidFileName":"/photoprism/storage/photoprism.pid","Env":["A=1"]}`+"\n"))
		require.NoError(t, err)
		assert.Equal(t, "/photoprism/storage/photoprism.pid", name)
	})
	t.Run("NoName", func(t *testing.T) {
		_, err := daemonPidFileName(pipe(t, `{"LogFileName":"photoprism.log"}`))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no pid file name")
	})
	t.Run("Invalid", func(t *testing.T) {
		_, err := daemonPidFileName(pipe(t, "not json"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid daemon context")
	})
	t.Run("Oversized", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		t.Cleanup(func() { _ = r.Close() })
		go func() {
			_, _ = w.WriteString(`{"Env":["` + strings.Repeat("x", daemonContextMaxSize) + `"],"PidFileName":"/tmp/x.pid"}`)
			_ = w.Close()
		}()
		_, err = daemonPidFileName(r)
		assert.Error(t, err)
	})
	t.Run("NotPipe", func(t *testing.T) {
		f, err := os.Open(unlockedPidFile(t, 1234))
		require.NoError(t, err)
		t.Cleanup(func() { _ = f.Close() })
		_, err = daemonPidFileName(f)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a pipe")
	})
}

// TestWaitForDaemonPid covers waiting for a daemon to write its process ID to the pid file.
func TestWaitForDaemonPid(t *testing.T) {
	t.Run("Written", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.WriteFile(pidFile, nil, 0o600))
		before := fileInfo(t, pidFile)
		time.AfterFunc(200*time.Millisecond, func() { _ = os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0o600) })
		require.NoError(t, waitForDaemonPid(pidFile, before, cmd.Process, 10*time.Second))
	})
	t.Run("Created", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		time.AfterFunc(200*time.Millisecond, func() { _ = os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0o600) })
		require.NoError(t, waitForDaemonPid(pidFile, nil, cmd.Process, 10*time.Second))
	})
	t.Run("Unchanged", func(t *testing.T) {
		// The pid file already names the process, as after process ID reuse.
		cmd := startProcess(t, "photoprism-test", "30")
		pidFile := unlockedPidFile(t, cmd.Process.Pid)
		err := waitForDaemonPid(pidFile, fileInfo(t, pidFile), cmd.Process, 300*time.Millisecond)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has not written its process id")
	})
	t.Run("RewrittenSame", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		pidFile := unlockedPidFile(t, cmd.Process.Pid)
		before := fileInfo(t, pidFile)
		time.AfterFunc(200*time.Millisecond, func() { _ = os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0o600) })
		require.NoError(t, waitForDaemonPid(pidFile, before, cmd.Process, 10*time.Second))
	})
	t.Run("Exited", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "0.2")
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		err := waitForDaemonPid(pidFile, nil, cmd.Process, 10*time.Second)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exited before writing its process id")
	})
	t.Run("OtherPid", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		pidFile := unlockedPidFile(t, cmd.Process.Pid+1)
		err := waitForDaemonPid(pidFile, nil, cmd.Process, 300*time.Millisecond)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has not written its process id")
	})
}

// TestPidFileChanged covers telling whether a pid file was written since it was last seen.
func TestPidFileChanged(t *testing.T) {
	t.Run("Created", func(t *testing.T) {
		assert.True(t, pidFileChanged(nil, fileInfo(t, unlockedPidFile(t, 1234))))
	})
	t.Run("Missing", func(t *testing.T) {
		assert.False(t, pidFileChanged(nil, nil))
		assert.False(t, pidFileChanged(fileInfo(t, unlockedPidFile(t, 1234)), nil))
	})
	t.Run("Same", func(t *testing.T) {
		info := fileInfo(t, unlockedPidFile(t, 1234))
		assert.False(t, pidFileChanged(info, info))
	})
	t.Run("Replaced", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 1234)
		before := keepFile(t, pidFile)
		replacePidFile(t, pidFile, "1234")
		assert.True(t, pidFileChanged(before, fileInfo(t, pidFile)))
	})
	t.Run("Resized", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 1234)
		before := fileInfo(t, pidFile)
		require.NoError(t, os.Chtimes(pidFile, before.ModTime(), before.ModTime()))
		require.NoError(t, os.Truncate(pidFile, 2))
		require.NoError(t, os.Chtimes(pidFile, before.ModTime(), before.ModTime()))
		assert.True(t, pidFileChanged(before, fileInfo(t, pidFile)))
	})
	t.Run("Modified", func(t *testing.T) {
		pidFile := unlockedPidFile(t, 1234)
		before := fileInfo(t, pidFile)
		later := before.ModTime().Add(time.Second)
		require.NoError(t, os.Chtimes(pidFile, later, later))
		assert.True(t, pidFileChanged(before, fileInfo(t, pidFile)))
	})
}

// TestDetachRequested covers deciding whether to start the daemon, and removing the daemon marker otherwise.
func TestDetachRequested(t *testing.T) {
	t.Run("Parent", func(t *testing.T) {
		t.Setenv(daemon.MARK_NAME, "")
		require.NoError(t, os.Unsetenv(daemon.MARK_NAME))
		assert.True(t, detachRequested(true))
	})
	t.Run("Daemon", func(t *testing.T) {
		t.Setenv(daemon.MARK_NAME, daemon.MARK_VALUE)
		assert.False(t, detachRequested(true))
		assert.False(t, daemon.WasReborn())
	})
	t.Run("Foreground", func(t *testing.T) {
		t.Setenv(daemon.MARK_NAME, daemon.MARK_VALUE)
		assert.False(t, detachRequested(false))
		assert.False(t, daemon.WasReborn())
	})
}

// TestDetachDaemon covers starting the daemon with go-daemon and waiting until it has written its process ID.
func TestDetachDaemon(t *testing.T) {
	// rebornContext returns a context that starts this test binary as a daemon in the reborn helper mode.
	rebornContext := func(t *testing.T) *daemon.Context {
		dir := t.TempDir()
		pidFile := filepath.Join(dir, "photoprism.pid")
		logFile := filepath.Join(dir, "photoprism.log")

		return &daemon.Context{
			PidFileName: pidFile,
			LogFileName: logFile,
			Args:        []string{"photoprism-test", "-test.run=^$"},
			Env:         append(os.Environ(), stopHelperEnv+"=reborn", stopHelperPidEnv+"="+pidFile, stopHelperLogEnv+"="+logFile),
		}
	}

	t.Run("Started", func(t *testing.T) {
		requireProc(t)
		dctx := rebornContext(t)
		pid, status, err := detachDaemon(dctx)
		require.NoError(t, err)

		child, findErr := os.FindProcess(pid)
		require.NoError(t, findErr)
		t.Cleanup(func() { _ = child.Kill(); _, _ = child.Wait() })

		assert.Equal(t, daemonStarted, status)
		// The daemon has written its process ID when detachDaemon returns.
		assertPidFileContent(t, dctx.PidFileName, strconv.Itoa(pid))
	})
	t.Run("Starting", func(t *testing.T) {
		// A stand-in daemon that is starting holds the lock and writes its process ID later.
		dctx := rebornContext(t)
		lock, err := daemon.OpenLockFile(dctx.PidFileName, daemon.FILE_PERM)
		require.NoError(t, err)
		require.NoError(t, lock.Lock())
		cmd := startProcess(t, "photoprism-test", "30", lock.File)
		require.NoError(t, lock.Close())
		time.AfterFunc(300*time.Millisecond, func() {
			_ = os.WriteFile(dctx.PidFileName, []byte(strconv.Itoa(cmd.Process.Pid)), daemon.FILE_PERM)
		})

		pid, status, err := detachDaemon(dctx)
		require.NoError(t, err)
		assert.Equal(t, daemonAlreadyRunning, status)
		assert.Equal(t, cmd.Process.Pid, pid)
	})
	t.Run("LockedWithoutDaemon", func(t *testing.T) {
		defaultTimeout := daemonPidTimeout
		daemonPidTimeout = 300 * time.Millisecond
		t.Cleanup(func() { daemonPidTimeout = defaultTimeout })

		// Another process holds the lock, but no daemon starts.
		dctx := rebornContext(t)
		lock, err := daemon.OpenLockFile(dctx.PidFileName, daemon.FILE_PERM)
		require.NoError(t, err)
		require.NoError(t, lock.Lock())
		t.Cleanup(func() { _ = lock.Close() })

		_, _, err = detachDaemon(dctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "is locked by another process, but no daemon holds it after 300ms")
		assert.Contains(t, err.Error(), "fuser -v")
		assertPidFileContent(t, dctx.PidFileName, "")
	})
	t.Run("Released", func(t *testing.T) {
		requireProc(t)
		// Another process holds the lock briefly and releases it without starting a daemon.
		dctx := rebornContext(t)
		lock, err := daemon.OpenLockFile(dctx.PidFileName, daemon.FILE_PERM)
		require.NoError(t, err)
		require.NoError(t, lock.Lock())
		time.AfterFunc(200*time.Millisecond, func() { _ = lock.Close() })

		pid, status, err := detachDaemon(dctx)
		require.NoError(t, err)

		child, findErr := os.FindProcess(pid)
		require.NoError(t, findErr)
		t.Cleanup(func() { _ = child.Kill(); _, _ = child.Wait() })

		assert.Equal(t, daemonStarted, status)
		assertPidFileContent(t, dctx.PidFileName, strconv.Itoa(pid))
	})
	t.Run("Flapping", func(t *testing.T) {
		defaultTimeout := daemonPidTimeout
		daemonPidTimeout = 500 * time.Millisecond
		t.Cleanup(func() { daemonPidTimeout = defaultTimeout })

		// The lock is released for every probe and held again for every start, so no daemon can start.
		dctx := rebornContext(t)
		lock, err := daemon.OpenLockFile(dctx.PidFileName, daemon.FILE_PERM)
		require.NoError(t, err)
		require.NoError(t, lock.Lock())
		t.Cleanup(func() { _ = lock.Close() })

		defaultProbe := probePidFileLock
		probes := 0
		probePidFileLock = func(*os.File) error { probes++; return nil }
		t.Cleanup(func() { probePidFileLock = defaultProbe })

		start := time.Now()
		_, _, err = detachDaemon(dctx)
		require.Error(t, err)
		assert.Less(t, time.Since(start), 5*time.Second)
		// Every retry follows a pause, so the number of attempts is bounded by the timeout.
		assert.LessOrEqual(t, probes, 10)
	})
	t.Run("StartingUnconfirmed", func(t *testing.T) {
		defaultTimeout, defaultHoldsFile, defaultProbe := daemonPidTimeout, holdsFile, probePidFileLock
		daemonPidTimeout = 300 * time.Millisecond
		holdsFile = func(int, string, os.FileInfo) (bool, bool) { return false, false }
		probePidFileLock = func(*os.File) error { return syscall.ENOLCK }
		t.Cleanup(func() { daemonPidTimeout, holdsFile, probePidFileLock = defaultTimeout, defaultHoldsFile, defaultProbe })

		// The starting stand-in holds the lock and writes its process ID later, but it cannot be confirmed.
		dctx := rebornContext(t)
		lock, err := daemon.OpenLockFile(dctx.PidFileName, daemon.FILE_PERM)
		require.NoError(t, err)
		require.NoError(t, lock.Lock())
		cmd := startProcess(t, "photoprism-test", "30", lock.File)
		require.NoError(t, lock.Close())
		time.AfterFunc(50*time.Millisecond, func() {
			_ = os.WriteFile(dctx.PidFileName, []byte(strconv.Itoa(cmd.Process.Pid)), daemon.FILE_PERM)
		})

		start := time.Now()
		_, _, err = detachDaemon(dctx)
		assert.ErrorIs(t, err, errDaemonUnconfirmed)
		assert.GreaterOrEqual(t, time.Since(start), 300*time.Millisecond)
	})
	t.Run("AlreadyRunning", func(t *testing.T) {
		cmd, pidFile := startDaemon(t, "30")
		dctx := rebornContext(t)
		dctx.PidFileName = pidFile
		pid, status, err := detachDaemon(dctx)
		require.NoError(t, err)
		assert.Equal(t, daemonAlreadyRunning, status)
		assert.Equal(t, cmd.Process.Pid, pid)
	})
	t.Run("Unreadable", func(t *testing.T) {
		dctx := rebornContext(t)
		require.NoError(t, os.Mkdir(dctx.PidFileName, 0o700))
		_, _, err := detachDaemon(dctx)
		assert.Error(t, err)
	})
}

// TestClearDaemonMark covers keeping the daemon marker out of the programs the server runs.
func TestClearDaemonMark(t *testing.T) {
	t.Run("Reborn", func(t *testing.T) {
		t.Setenv(daemon.MARK_NAME, daemon.MARK_VALUE)
		require.True(t, daemon.WasReborn())
		clearDaemonMark()
		assert.False(t, daemon.WasReborn())

		out, err := exec.Command("env").Output() //nolint:gosec // Runs env from PATH.
		require.NoError(t, err)
		assert.NotContains(t, string(out), daemon.MARK_NAME+"=")
	})
	t.Run("NotSet", func(t *testing.T) {
		t.Setenv(daemon.MARK_NAME, "")
		require.NoError(t, os.Unsetenv(daemon.MARK_NAME))
		clearDaemonMark()
		_, set := os.LookupEnv(daemon.MARK_NAME)
		assert.False(t, set)
	})
}
