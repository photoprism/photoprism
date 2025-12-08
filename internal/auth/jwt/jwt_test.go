package jwt

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	cfg "github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {
	// Init test logger.
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	// Run unit tests.
	code := m.Run()

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

	os.Exit(code)
}

func newTestConfig(t *testing.T) *cfg.Config {
	return cfg.NewMinimalTestConfig(t.TempDir())
}
