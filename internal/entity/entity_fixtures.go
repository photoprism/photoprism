package entity

import (
	"time"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/clean"
)

// CreateDefaultFixtures inserts default fixtures for test and production.
func CreateDefaultFixtures() {
	CreateDefaultUsers()
	CreateUnknownPlace()
	CreateUnknownLocation()
	CreateUnknownCountry()
	CreateUnknownCamera()
	CreateUnknownLens()

	// Default fixtures have fixed IDs.
	if err := SyncSequences(Db()); err != nil {
		log.Warnf("migrate: %s", clean.Error(err))
	}
}

// hasDefaultFixtureTables checks if the tables that CreateDefaultFixtures reads and writes exist.
func hasDefaultFixtureTables(db *gorm.DB) bool {
	for _, name := range []string{User{}.TableName(), Place{}.TableName(), Cell{}.TableName(),
		Country{}.TableName(), Camera{}.TableName(), Lens{}.TableName()} {
		if found, err := DbHasTable(db, name); err != nil || !found {
			return false
		}
	}

	return true
}

// ResetTestFixtures recreates database tables and test fixtures.
func ResetTestFixtures() {
	start := time.Now()

	// Make sure that the migrations and versions tables are already there, as once prevents these from being handled correctly in tests.
	if !Db().Migrator().HasTable(&migrate.Migration{}) {
		Log("migrate", "create migrations table", Db().AutoMigrate(&migrate.Migration{}))
	}
	if !Db().Migrator().HasTable(&migrate.Version{}) {
		Log("migrate", "create versions table", Db().AutoMigrate(&migrate.Version{}))
	}

	if err := Entities.Migrate(Db(), migrate.Opt(true, false, nil)); err != nil {
		log.Errorf("migrate: %s [%s]", clean.Error(err), time.Since(start))
	}

	if err := Entities.WaitForMigration(Db()); err != nil {
		log.Errorf("migrate: %s [%s]", clean.Error(err), time.Since(start))
	}

	Entities.Truncate(Db())

	CreateDefaultFixtures()

	CreateTestFixtures()

	if err := SyncSequences(Db()); err != nil {
		log.Errorf("migrate: %s [%s]", clean.Error(err), time.Since(start))
	}

	FlushCaches()

	File{}.RegenerateIndex()

	log.Debugf("migrate: recreated test fixtures [%s]", time.Since(start))
}
