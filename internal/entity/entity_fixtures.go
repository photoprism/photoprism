package entity

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/entity/migrate"
)

// CreateDefaultFixtures inserts default fixtures for test and production.
func CreateDefaultFixtures() {
	CreateDefaultUsers()
	CreateUnknownPlace()
	CreateUnknownLocation()
	CreateUnknownCountry()
	CreateUnknownCamera()
	CreateUnknownLens()
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
	if (!Db().HasTable(&migrate.Migration{})) {
		Db().AutoMigrate(&migrate.Migration{})
	}
	if (!Db().HasTable(&migrate.Version{})) {
		Db().AutoMigrate(&migrate.Version{})
	}

	if err := Entities.Migrate(Db(), migrate.Opt(true, false, nil)); err != nil {
		log.Errorf("migrate: %s [%s]", err, time.Since(start))
	}

	if err := Entities.WaitForMigration(Db()); err != nil {
		log.Errorf("migrate: %s [%s]", err, time.Since(start))
	}

	Entities.Truncate(Db())

	CreateDefaultFixtures()

	CreateTestFixtures()

	FlushCaches()

	File{}.RegenerateIndex()

	log.Debugf("migrate: recreated test fixtures [%s]", time.Since(start))
}
