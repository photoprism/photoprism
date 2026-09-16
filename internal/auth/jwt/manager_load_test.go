package jwt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestManagerLoadKeysUnreadableFile(t *testing.T) {
	c := newTestConfig(t)
	keyDir := filepath.Join(c.PortalConfigPath(), "keys")
	require.NoError(t, os.MkdirAll(keyDir, 0o700))

	// A dangling link is listed like a key file and fails to open whatever the process owns.
	broken := filepath.Join(keyDir, privateKeyPrefix+"20250924T1030Z-broken"+privateKeyExt)
	require.NoError(t, os.Symlink(filepath.Join(keyDir, "missing.json"), broken))

	logger, ok := log.(*logrus.Logger)
	require.True(t, ok)
	hook := logtest.NewLocal(logger)
	t.Cleanup(hook.Reset)

	m, err := NewManager(c)
	require.NoError(t, err)
	require.NotNil(t, m)

	entries := hook.AllEntries()
	require.NotEmpty(t, entries)

	var warned bool

	for _, entry := range entries {
		msg := entry.Message
		if entry.Level != logrus.WarnLevel {
			continue
		}
		warned = true
		require.NotContains(t, msg, keyDir)
		require.NotContains(t, msg, c.PortalConfigPath())
		require.Contains(t, msg, "read signing key")
		// The file name stays, so an operator can still tell which entry was skipped.
		require.Contains(t, msg, "broken")
	}

	require.True(t, warned, "expected a warning for the unreadable key file")
}
