package jwt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/event"
)

func TestManagerLoadKeysUnreadableFile(t *testing.T) {
	c := newTestConfig(t)
	keyDir := filepath.Join(c.PortalConfigPath(), "keys")
	require.NoError(t, os.MkdirAll(keyDir, 0o700))

	// A dangling link is listed like a key file and fails to open whatever the process owns.
	broken := filepath.Join(keyDir, privateKeyPrefix+"20250924T1030Z-broken"+privateKeyExt)
	require.NoError(t, os.Symlink(filepath.Join(keyDir, "missing.json"), broken))

	// Operator-only detail goes to the system log.
	logger, hook := logtest.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	prev := event.SystemLog
	event.SystemLog = logger
	t.Cleanup(func() { event.SystemLog = prev })

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
		require.Contains(t, msg, "read signing key")
		// The console keeps the location, so an operator can tell which entry was skipped and where.
		require.Contains(t, msg, keyDir)
		require.Contains(t, msg, "broken")
	}

	require.True(t, warned, "expected a warning for the unreadable key file")
}
