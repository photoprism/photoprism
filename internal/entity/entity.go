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

type Table struct {
	Field    string
	Type     string
	Null     string
	Key      string
	Default  string
	Extra    string
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

	WaitForMigration()

	CreateUnknownPlace()
	CreateUnknownCountry()
	CreateUnknownCamera()
	CreateUnknownLens()
}

// Waits for tables to be available after migrating / resetting database.
func WaitForMigration() {
	for i := 0; i < 20; i++ {
		table := Table{}

		if Db().Raw("DESCRIBE places").Scan(&table).Error == nil {
			return
		}

		time.Sleep(50 * time.Millisecond)
	}
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
	MigrateDb()

	if testFixtures {
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
