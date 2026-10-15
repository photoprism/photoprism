package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
		null, err := os.Open(os.DevNull)
		require.NoError(t, err)
		t.Cleanup(func() { _ = null.Close() })

		context, err := json.Marshal(daemonContext{PidFileName: pidFile})
		require.NoError(t, err)

		cmd, stdout := startHelperInput(t, "write-pid", []string{"_GO_DAEMON=1"}, bytes.NewReader(context), null, f)
		line, err := stdout.ReadString('\n')
		require.NoError(t, err)
		require.NoError(t, cmd.Wait())

		pidText, result, found := strings.Cut(strings.TrimPrefix(strings.TrimSpace(line), "pid="), " err=")
		require.True(t, found, line)
		pid, err := strconv.Atoi(pidText)
		require.NoError(t, err, line)

		return pid, result
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
		time.AfterFunc(200*time.Millisecond, func() { _ = os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0o600) })
		require.NoError(t, waitForDaemonPid(pidFile, cmd.Process, 10*time.Second))
	})
	t.Run("Exited", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "0.2")
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		err := waitForDaemonPid(pidFile, cmd.Process, 10*time.Second)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exited before writing its process id")
	})
	t.Run("OtherPid", func(t *testing.T) {
		cmd := startProcess(t, "photoprism-test", "30")
		pidFile := unlockedPidFile(t, cmd.Process.Pid+1)
		err := waitForDaemonPid(pidFile, cmd.Process, 300*time.Millisecond)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has not written its process id")
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
