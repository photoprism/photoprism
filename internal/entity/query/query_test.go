package query

import (
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

// staticDbProvider returns a static *gorm.DB for temporary test provider overrides.
type staticDbProvider struct {
	db *gorm.DB
}

// Db returns the static database handle.
func (p staticDbProvider) Db() *gorm.DB {
	return p.db
}

// testDriver returns the driver the test database runs on, applying the same
// fallback to SQLite that entity.InitTestDb uses when resolving the environment.
func testDriver() string {
	switch driver := os.Getenv("PHOTOPRISM_TEST_DRIVER"); {
	case os.Getenv("PHOTOPRISM_TEST_DSN") == "", driver == "", driver == "test", driver == "sqlite":
		return dsn.DriverSQLite3
	default:
		return driver
	}
}

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

	db := entity.InitTestDb(
		os.Getenv("PHOTOPRISM_TEST_DRIVER"),
		os.Getenv("PHOTOPRISM_TEST_DSN"))
	defer db.Close()

	return m.Run()
}

func TestDbDialect(t *testing.T) {
	t.Run("TestDriver", func(t *testing.T) {
		assert.Equal(t, testDriver(), DbDialect())
	})
}

func TestBatchSize(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) {
		if testDriver() != dsn.DriverSQLite3 {
			t.Skip("test database is not SQLite")
		}
		assert.Equal(t, 333, BatchSize())
	})
	t.Run("MySQL", func(t *testing.T) {
		if testDriver() != dsn.DriverMySQL {
			t.Skip("test database is not MySQL")
		}
		assert.Equal(t, 1000, BatchSize())
	})
}
