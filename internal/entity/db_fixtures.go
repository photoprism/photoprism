package entity

import (
	"time"
)

// CreateDefaultFixtures inserts default fixtures for test and production.
func CreateDefaultFixtures() {
	CreateUnknownAddress()
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

	Entities.Migrate(Db(), false, nil)
	Entities.WaitForMigration(Db())
	Entities.Truncate(Db())

	CreateDefaultFixtures()

	CreateTestFixtures()

	log.Debugf("entity: recreated test fixtures [%s]", time.Since(start))
}
