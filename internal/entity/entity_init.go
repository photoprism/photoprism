package entity

import (
	"os"
	"time"

	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// onReady stores callbacks to execute once database initialization finishes.
var onReady []func()

// ready executes all registered callbacks after database initialization completes.
func ready() {
	for _, init := range onReady {
		init()
	}
}

// InitDb creates database tables and inserts default fixtures as needed.
func InitDb(opt migrate.Options) {
	if !HasDbProvider() {
		log.Error("migrate: no database provider")
		return
	}

	start := time.Now()

	if opt.DropDeprecated {
		DeprecatedTables.Drop(Db())
	}

	Entities.Migrate(Db(), opt)

	if err := Entities.WaitForMigration(Db()); err != nil {
		log.Errorf("migrate: %s", err)
	}

	CreateDefaultFixtures()

	ready()

	log.Debugf("migrate: completed in %s", time.Since(start))
}

// InitTestDb connects to and completely initializes the test database incl fixtures.
func InitTestDb(driver, dbDsn string) *DbConn {
	if HasDbProvider() {
		return nil
	}

	start := time.Now()

	// Set default test database driver.
	if driver == "test" || driver == "sqlite" || driver == "" || dbDsn == "" {
		driver = SQLite3
	}

	// Set default database DSN.
	if driver == SQLite3 {
		if dbDsn == "" || dbDsn == SQLiteTestDB {
			dbDsn = SQLiteTestDB
			if !fs.FileExists(dbDsn) {
				log.Debugf("sqlite: test database %s does not already exist", clean.Log(dbDsn))
			} else if err := os.Remove(dbDsn); err != nil {
				log.Errorf("sqlite: failed to remove existing test database %s (%s)", clean.Log(dbDsn), err)
			}
		}
	}

	log.Infof("initializing %s test db in %s", driver, dbDsn)

	// Create gorm.DB connection provider.
	db := &DbConn{
		Driver: driver,
		Dsn:    dbDsn,
	}

	// Insert test fixtures into the database.
	SetDbProvider(db)
	ResetTestFixtures()
	File{}.RegenerateIndex()

	ready()

	log.Debugf("migrate: completed in %s", time.Since(start))

	return db
}
