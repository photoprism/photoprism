package entity

import (
	"os"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/testextras"
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
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) (code int) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	driver, dsname := dsn.PhotoPrismTestToDriverDSN()

	db := entity.InitTestDb(
		driver,
		dsname)

	defer db.Close()
	// Run unit tests.
	return testextras.TestDbCleanup(m.Run())
}
