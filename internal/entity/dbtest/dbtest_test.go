package entity

import (
	"os"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/clean"
)

var log = event.Log

// All tests in this suite MUST lock and unlock this mutex or they will fail
// on SQLite which doesn't support row locking.
var dbtestMutex = sync.Mutex{}

// Log logs the error if any and keeps quiet otherwise.
func Log(model, action string, err error) {
	if err != nil {
		log.Errorf("%s: %s (%s)", model, err, action)
	}
}

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	caller := "internal/entity/dbtest/dbtest_test.go/TestMain"
	dbc, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	driver := os.Getenv("PHOTOPRISM_TEST_DRIVER")
	dsn := os.Getenv("PHOTOPRISM_TEST_DSN")

	// Set default test database driver.
	if driver == "test" || driver == "sqlite" || driver == "" || dsn == "" {
		driver = entity.SQLite3
	}

	// Set default database DSN.
	if driver == entity.SQLite3 {
		if dsn == "" {
			dsn = entity.SQLiteMemoryDSN
		} else if dsn != entity.SQLiteTestDB {
			// Continue.
		} else if err := os.Remove(dsn); err == nil {
			log.Debugf("sqlite: test file %s removed", clean.Log(dsn))
		}
	}

	db := entity.InitTestDb(
		driver,
		dsn)

	defer db.Close()

	code := m.Run()

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	os.Exit(code)
}
