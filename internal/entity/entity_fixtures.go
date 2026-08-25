package entity

import (
	"time"

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
}

// ResetTestFixtures recreates or truncates database tables and recreates default and test fixtures.
func ResetTestFixtures() {
	start := time.Now()
	runMigration := true

	// Make sure that the migrations and versions tables are already there, as once prevents these from being handled correctly in tests.
	if (!Db().Migrator().HasTable(&migrate.Migration{})) {
		if err := Db().Migrator().AutoMigrate(&migrate.Migration{}); err != nil {
			log.Errorf("migrate: automigrate of migrations - aborting (%s)", clean.Log(err.Error()))
			return
		}
	}
	if (!Db().Migrator().HasTable(&migrate.Version{})) {
		if err := Db().Migrator().AutoMigrate(&migrate.Version{}); err != nil {
			log.Errorf("migrate: automigrate of versions - aborting (%s)", clean.Log(err.Error()))
			return
		}
	}

	version := migrate.NewVersion("DBMS AuthID Fix", "Any Editions")
	if version.Find(Db()) != nil {
		runMigration = false
		log.Debugf("migrate: skipping table migration as version %s was found from %s", version.Version, version.CreatedAt)
	}

	// Only run the migration if the DBMS AuthID Fix is not there.  It's existence indicates that migration has already been run.
	if runMigration {
		Entities.Migrate(Db(), migrate.Opt(true, false, nil))

		if err := Entities.WaitForMigration(Db()); err != nil {
			log.Errorf("migrate: %s [%s]", err, time.Since(start))
		}
	} else {
		Entities.Truncate(Db())
	}

	CreateDefaultFixtures()

	CreateTestFixtures()

	FlushCaches()

	File{}.RegenerateIndex()

	log.Debugf("migrate: recreated test fixtures [%s]", time.Since(start))
}

// ResetNoTestFixtures recreates or truncates database tables and recreates default fixtures.
func ResetNoTestFixtures() {
	start := time.Now()
	runMigration := true

	// Make sure that the migrations and versions tables are already there, as once prevents these from being handled correctly in tests.
	if (!Db().Migrator().HasTable(&migrate.Migration{})) {
		if err := Db().Migrator().AutoMigrate(&migrate.Migration{}); err != nil {
			log.Errorf("migrate: automigrate of migrations - aborting (%s)", clean.Log(err.Error()))
			return
		}
	}
	if (!Db().Migrator().HasTable(&migrate.Version{})) {
		if err := Db().Migrator().AutoMigrate(&migrate.Version{}); err != nil {
			log.Errorf("migrate: automigrate of versions - aborting (%s)", clean.Log(err.Error()))
			return
		}
	}

	version := migrate.NewVersion("DBMS AuthID Fix", "Any Editions")
	if version.Find(Db()) != nil {
		runMigration = false
		log.Debugf("migrate: skipping table migration as version %s was found from %s", version.Version, version.CreatedAt)
	}

	// Only run the migration if the DBMS AuthID Fix is not there.  It's existence indicates that migration has already been run.
	if runMigration {
		Entities.Migrate(Db(), migrate.Opt(true, false, nil))

		if err := Entities.WaitForMigration(Db()); err != nil {
			log.Errorf("migrate: %s [%s]", err, time.Since(start))
		}
	}

	Entities.Truncate(Db())

	CreateDefaultFixtures()

	FlushCaches()

	File{}.RegenerateIndex()

	log.Debugf("migrate: recreated default fixtures [%s]", time.Since(start))
}
