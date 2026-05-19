package session

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	c := config.TestConfig()
	defer c.CleanupTestFolder()
	defer func() {
		if err := c.CloseDb(); err != nil {
			log.Warnf("close db: %v", err)
		}
		// Remove temporary SQLite files after running the tests.
		fs.PurgeTestDbFiles(".", false)
	}()

	return m.Run()
}
