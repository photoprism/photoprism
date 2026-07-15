package jwt

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	cfg "github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	// Init test logger.
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)

	// Remove temporary SQLite files after running the tests.
	defer fs.PurgeTestDbFiles(".", false)

	// Run unit tests.
	return m.Run()
}

func newTestConfig(t *testing.T) *cfg.Config {
	return cfg.NewMinimalTestConfig("", t.TempDir())
}
