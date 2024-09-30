package entity

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
)

var log = event.Log

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

	os.Exit(code)
}
