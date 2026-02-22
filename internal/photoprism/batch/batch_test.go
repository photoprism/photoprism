package batch

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestMain executes testMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

// testMain configures shared state for the batch action tests.
func testMain(m *testing.M) (code int) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)

	c := config.TestConfig()
	defer c.CleanupTestFolder()
	defer func() {
		// Prevent UpdateCountsAsync from causing the test suite to fail due to the database closing before the goroutine has finished.
		mutex.Index.Lock()
		if err := c.CloseDb(); err != nil {
			log.Warnf("close db: %v", err)
		}
		// Remove temporary SQLite files after running the tests.
		fs.PurgeTestDbFiles(".", false)
	}()

	get.SetConfig(c)
	photoprism.SetConfig(c)
	return m.Run()
}
