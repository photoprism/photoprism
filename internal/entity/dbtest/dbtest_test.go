package entity

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
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
	os.Exit(testMain(m))
}

func testMain(m *testing.M) (code int) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	caller := "internal/entity/dbtest/dbtest_test.go/TestMain"
	dbc, dbn, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		return 1
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	driver, dsname := dsn.PhotoPrismTestToDriverDSN(dbn)
	dsn.SetDSNToEnv(dsname)

	// Set default test database driver.
	if driver == "test" || driver == "sqlite" || driver == "" || dsname == "" {
		driver = dsn.DriverSQLite3
	}

	// Set default database DSN.
	if driver == dsn.DriverSQLite3 {
		switch dsname {
		case "":
			dsname = dsn.SQLiteMemoryShared
		case dsn.SQLiteTestDB:
			if err := os.Remove(dsname); err == nil {
				log.Debugf("sqlite: test file %s removed", clean.Log(dsname))
			}
		}
	}

	db := entity.InitTestDb(
		driver,
		dsname)

	defer db.Close()

	beforeTimestamp := time.Now().UTC()
	code = m.Run()
	code = testextras.ValidateDBErrors(db.Db(), log, beforeTimestamp, code)

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	return code
}
