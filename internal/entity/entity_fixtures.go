package entity

import (
	"time"

	"github.com/photoprism/photoprism/internal/migrate"
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

	Entities.Migrate(Db(), migrate.Opt(true, false, nil))
	Entities.WaitForMigration(Db())
	Entities.Truncate(Db())

	CreateDefaultFixtures()

	CreateTestFixtures()

	log.Debugf("migrate: recreated test fixtures [%s]", time.Since(start))
}
