package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dustin/go-humanize/english"
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
	t.Run("DanglingSymlink", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.Symlink(filepath.Join(t.TempDir(), "missing.pid"), pidFile))
		_, _, err := daemonRunning(pidFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be a symbolic link")
	})
}
