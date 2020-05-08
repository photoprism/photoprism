/*
Package entity contains models for data storage based on GORM.

See http://gorm.io/docs/ for more information about GORM.

Additional information concerning data storage can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki/Storage
*/
package entity

import (
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log
var resetFixturesOnce sync.Once

func logError(result *gorm.DB) {
	if result.Error != nil {
		log.Error(result.Error.Error())
	}
}

// MigrateDb creates all tables and inserts default entities as needed.
func MigrateDb() {
	Db().AutoMigrate(
		&Account{},
		&File{},
		&FileShare{},
		&FileSync{},
		&Photo{},
		&Description{},
		&Place{},
		&Location{},
		&Camera{},
		&Lens{},
		&Country{},
		&Album{},
		&PhotoAlbum{},
		&Label{},
		&Category{},
		&PhotoLabel{},
		&Keyword{},
		&PhotoKeyword{},
		&Link{},
	)

	CreateUnknownPlace()
	CreateUnknownCountry()
	CreateUnknownCamera()
	CreateUnknownLens()
}

// DropTables drops database tables for all known entities.
func DropTables() {
	Db().DropTableIfExists(
		&Account{},
		&File{},
		&FileShare{},
		&FileSync{},
		&Photo{},
		&Description{},
		&Place{},
		&Location{},
		&Camera{},
		&Lens{},
		&Country{},
		&Album{},
		&PhotoAlbum{},
		&Label{},
		&Category{},
		&PhotoLabel{},
		&Keyword{},
		&PhotoKeyword{},
		&Link{},
	)
}

// ResetDb drops database tables for all known entities and re-creates them with fixtures.
func ResetDb(testFixtures bool) {
	DropTables()

	// Make sure changes have been written to disk.
	time.Sleep(200 * time.Millisecond)

	MigrateDb()

	if testFixtures {
		// Make sure changes have been written to disk.
		time.Sleep(200 * time.Millisecond)

		CreateTestFixtures()
	}
}

// InitTestFixtures resets the database and test fixtures once.
func InitTestFixtures() {
	resetFixturesOnce.Do(func() {
		ResetDb(true)
	})
}

// InitTestDb connects to and completely initializes the test database incl fixtures.
func InitTestDb(dsn string) *Gorm {
	if HasDbProvider() {
		return nil
	}

	db := &Gorm{
		Driver: "mysql",
		Dsn:    dsn,
	}

	SetDbProvider(db)
	InitTestFixtures()

	return db
}
