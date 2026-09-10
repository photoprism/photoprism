package migrate

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) (code int) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)
	// Remove temporary SQLite files after running the tests.
	defer fs.PurgeTestDbFiles(".", false)

	return testextras.TestDbCleanup(m.Run())
}
