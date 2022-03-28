package entity

import (
	"os"
	"time"

	"github.com/photoprism/photoprism/pkg/sanitize"
)

// MigrateDb creates database tables and inserts default fixtures as needed.
func MigrateDb(dropDeprecated, runFailed bool) {
	start := time.Now()

	if dropDeprecated {
		DeprecatedTables.Drop(Db())
	}

	Entities.Migrate(Db(), runFailed)
	Entities.WaitForMigration(Db())

	CreateDefaultFixtures()

	log.Debugf("entity: successfully initialized [%s]", time.Since(start))
}

// InitTestDb connects to and completely initializes the test database incl fixtures.
func InitTestDb(driver, dsn string) *Gorm {
	if HasDbProvider() {
		return nil
	}

	// Set default test database driver.
	if driver == "test" || driver == "sqlite" || driver == "" || dsn == "" {
		driver = SQLite3
	}

	// Set default database DSN.
	if driver == SQLite3 {
		if dsn == "" {
			dsn = SQLiteMemoryDSN
		} else if dsn != SQLiteTestDB {
			// Continue.
		} else if err := os.Remove(dsn); err == nil {
			log.Debugf("sqlite: test file %s removed", sanitize.Log(dsn))
		}
	}

	log.Infof("initializing %s test db in %s", driver, dsn)

	// Create ORM instance.
	db := &Gorm{
		Driver: driver,
		Dsn:    dsn,
	}

	// Insert test fixtures.
	SetDbProvider(db)
	ResetTestFixtures()

	return db
}
