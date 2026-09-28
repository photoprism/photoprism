package backup

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/txt"
)

// captureLog replaces the package logger for the duration of a test and returns its entries.
func captureLog(t *testing.T) *test.Hook {
	t.Helper()

	orig := log
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	log = logger

	t.Cleanup(func() { log = orig })

	return hook
}

// TestDatabase_CommandTrace checks how a database backup renders its client command.
// RestoreDatabase renders its own through the same helper; it is not run here because a restore
// would replace the test database.
func TestDatabase_CommandTrace(t *testing.T) {
	c := get.Config()

	switch c.DatabaseDriver() {
	case dsn.DriverMySQL, dsn.DriverMariaDB:
		// Only these drivers pass connection options to a client binary.
	default:
		t.Skip("requires a MySQL or MariaDB test database")
	}

	password := c.DatabasePassword()
	require.NotEmpty(t, password)
	require.NotEmpty(t, c.MariadbDumpBin(), "the client binary must be found for the trace to name it")

	hook := captureLog(t)
	backupPath := t.TempDir()

	require.NoError(t, Database(backupPath, filepath.Join(backupPath, "trace.sql"), false, true, 0))

	var traced string

	for _, entry := range hook.AllEntries() {
		assert.NotContains(t, entry.Message, password)

		if entry.Level == logrus.TraceLevel && strings.Contains(entry.Message, c.MariadbDumpBin()) {
			traced = entry.Message
		}
	}

	require.NotEmpty(t, traced, "expected the dump command to be traced")

	// The client reads the password from the environment, so the trace has none to mask.
	assert.NotContains(t, traced, "-p"+txt.Masked)
	assert.Contains(t, traced, "--no-defaults")

	// A client that can verify the zero-configuration TLS certificate is asked to.
	if major, minor, kind := mariadbClientVersion(c.MariadbDumpBin()); c.DatabaseSsl() && kind == clientMariadb && (major > 11 || major == 11 && minor >= 4) {
		assert.Contains(t, traced, "--ssl-verify-server-cert")
	}
}
