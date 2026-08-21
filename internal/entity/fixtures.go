package entity

import (
	"sync"

	"github.com/jinzhu/gorm"
)

var (
	fixtureMutex sync.RWMutex
	fixtureTx    *gorm.DB
)

// fixtureDb returns the open fixture transaction, or the default connection when there is
// none. Fixture inserts go through this instead of Db() so the transaction stays invisible
// to the background workers that read the shared connection provider.
func fixtureDb() *gorm.DB {
	fixtureMutex.RLock()
	defer fixtureMutex.RUnlock()

	if fixtureTx != nil {
		return fixtureTx
	}

	return Db()
}

// CreateTestFixtures inserts all known entities into the database for testing.
//
// The inserts share one transaction, since committing each of the several hundred rows
// separately costs an fsync per row on SQLite.
func CreateTestFixtures() {
	defer beginFixtureTx()()

	if err := Admin.SetPassword("photoprism"); err != nil {
		log.Error(err)
	}

	CreateLabelFixtures()
	CreateCameraFixtures()
	CreateCountryFixtures()
	CreatePhotoFixtures()
	CreateDetailsFixtures()
	CreateAlbumFixtures()
	CreateServiceFixtures()
	CreateLinkFixtures()
	CreatePhotoAlbumFixtures()
	CreateFolderFixtures()
	CreateFileFixtures()
	CreateKeywordFixtures()
	CreatePhotoKeywordFixtures()
	CreateCategoryFixtures()
	CreateCellFixtures()
	CreatePlaceFixtures()
	CreateFileShareFixtures()
	CreateFileSyncFixtures()
	CreateLensFixtures()
	CreateSubjectFixtures()
	CreateMarkerFixtures()
	CreateFaceFixtures()
	CreateUserFixtures()
	CreateSessionFixtures()
	CreateClientFixtures()
	CreateReactionFixtures()
	CreatePasscodeFixtures()
	CreatePasswordFixtures()
	CreateUserShareFixtures()
}

// beginFixtureTx opens a transaction for fixtureDb and returns the function that closes it:
// it rolls back and re-panics if the caller panicked, and otherwise commits. A commit that
// fails leaves no fixtures at all, so it is fatal rather than logged.
//
// Both are no-ops when a transaction is already open, when no provider is set, or when the
// transaction cannot be started; the inserts then run on whatever fixtureDb returns.
func beginFixtureTx() func() {
	if !HasDbProvider() {
		return func() {}
	}

	fixtureMutex.RLock()
	nested := fixtureTx != nil
	fixtureMutex.RUnlock()

	// Reuse the open transaction rather than starting a second writer against it.
	if nested {
		return func() {}
	}

	tx := Db().Begin()

	if tx.Error != nil {
		log.Warnf("fixtures: %s (begin transaction)", tx.Error)
		return func() {}
	}

	fixtureMutex.Lock()
	fixtureTx = tx
	fixtureMutex.Unlock()

	return func() {
		fixtureMutex.Lock()
		fixtureTx = nil
		fixtureMutex.Unlock()

		if r := recover(); r != nil {
			if err := tx.Rollback().Error; err != nil {
				log.Warnf("fixtures: %s (rollback)", err)
			}

			panic(r)
		}

		if err := tx.Commit().Error; err != nil {
			log.Fatalf("fixtures: %s (commit)", err)
		}
	}
}
