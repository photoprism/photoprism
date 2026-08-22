package entity

import (
	"time"

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

	Entities.Migrate(Db(), migrate.Opt(true, false, nil))

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
