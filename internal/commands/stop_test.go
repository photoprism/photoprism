package commands

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/sevlyar/go-daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// TestDaemonNotRunning covers which pid file search failures mean that no daemon is running.
func TestDaemonNotRunning(t *testing.T) {
	t.Run("NoPidFile", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		_, err := (&daemon.Context{PidFileName: pidFile}).Search()
		require.Error(t, err)
		assert.True(t, daemonNotRunning(err, pidFile))
	})
	t.Run("StalePidFile", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, fs.WriteString(pidFile, strconv.Itoa(3999999)))
		_, err := (&daemon.Context{PidFileName: pidFile}).Search()
		require.Error(t, err)
		assert.NoFileExists(t, pidFile)
		assert.True(t, daemonNotRunning(err, pidFile))
	})
	t.Run("UnreadablePidFile", func(t *testing.T) {
		pidFile := filepath.Join(t.TempDir(), "photoprism.pid")
		require.NoError(t, os.WriteFile(pidFile, []byte("not a pid"), fs.ModeFile))
		assert.False(t, daemonNotRunning(errors.New("daemon: invalid pid file"), pidFile))
	})
}
