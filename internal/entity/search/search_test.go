package search

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)
	// Remove temporary SQLite files after running the tests.
	defer fs.PurgeTestDbFiles(".", false)

	driver, dsname := dsn.PhotoPrismTestToDriverDSN()
	db := entity.InitTestDb(
		driver,
		dsname)
	defer db.Close()

	return testextras.TestDbCleanup(m.Run())
}

// testDialect returns the name of the SQL dialect the test database runs on, so
// that tests can account for collation and sort order differences.
func testDialect() string {
	return entity.Db().Dialector.Name()
}
