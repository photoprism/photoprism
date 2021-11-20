/*
Package entity contains models for data storage based on GORM.

See http://gorm.io/docs/ for more information about GORM.

Additional information concerning data storage can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki/Storage
*/
package entity

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log
var GeoApi = "places"

// logError logs the message if the argument is an error.
func logError(result *gorm.DB) {
	if result.Error != nil {
		log.Error(result.Error.Error())
	}
}

// TypeString returns an entity type string for logging.
func TypeString(entityType string) string {
	if entityType == "" {
		return "unknown"
	}

	return entityType
}

type Types map[string]interface{}

// Entities contains database entities and their table names.
var Entities = Types{
	"errors":              &Error{},
	"addresses":           &Address{},
	"users":               &User{},
	"accounts":            &Account{},
	"folders":             &Folder{},
	"duplicates":          &Duplicate{},
	"files":               &File{},
	"files_share":         &FileShare{},
	"files_sync":          &FileSync{},
	"photos":              &Photo{},
	"details":             &Details{},
	"places":              &Place{},
	"cells":               &Cell{},
	"cameras":             &Camera{},
	"lenses":              &Lens{},
	"countries":           &Country{},
	"albums":              &Album{},
	"photos_albums":       &PhotoAlbum{},
	"labels":              &Label{},
	"categories":          &Category{},
	"photos_labels":       &PhotoLabel{},
	"keywords":            &Keyword{},
	"photos_keywords":     &PhotoKeyword{},
	"passwords":           &Password{},
	"links":               &Link{},
	Subject{}.TableName(): &Subject{},
	Face{}.TableName():    &Face{},
	Marker{}.TableName():  &Marker{},
}

type RowCount struct {
	Count int
}

// WaitForMigration waits for the database migration to be successful.
func (list Types) WaitForMigration() {
	attempts := 100
	for name := range list {
		for i := 0; i <= attempts; i++ {
			count := RowCount{}
			if err := Db().Raw(fmt.Sprintf("SELECT COUNT(*) AS count FROM %s", name)).Scan(&count).Error; err == nil {
				log.Tracef("entity: %s migrated", txt.Quote(name))
				break
			} else {
				log.Debugf("entity: waiting for %s migration (%s)", txt.Quote(name), err.Error())
			}

			if i == attempts {
				panic("migration failed")
			}

			time.Sleep(50 * time.Millisecond)
		}
	}
}

// Truncate removes all data from tables without dropping them.
func (list Types) Truncate() {
	for name := range list {
		if err := Db().Exec(fmt.Sprintf("DELETE FROM %s WHERE 1", name)).Error; err == nil {
			// log.Debugf("entity: removed all data from %s", name)
			break
		} else if err.Error() != "record not found" {
			log.Debugf("entity: %s in %s", err, txt.Quote(name))
		}
	}
}

// Migrate migrates all database tables of registered entities.
func (list Types) Migrate() {
	for name, entity := range list {
		if err := UnscopedDb().AutoMigrate(entity).Error; err != nil {
			log.Debugf("entity: %s (waiting 1s)", err.Error())

			time.Sleep(time.Second)

			if err := UnscopedDb().AutoMigrate(entity).Error; err != nil {
				log.Errorf("entity: failed migrating %s", txt.Quote(name))
				panic(err)
			}
		}
	}
}

// Drop drops all database tables of registered entities.
func (list Types) Drop() {
	for _, entity := range list {
		if err := UnscopedDb().DropTableIfExists(entity).Error; err != nil {
			panic(err)
		}
	}
}

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

// MigrateIndexes runs additional table index migration queries.
func MigrateIndexes() {
	if err := Db().Exec("DROP INDEX IF EXISTS idx_places_place_label ON places").Error; err != nil {
		log.Errorf("%s: %s (drop index)", DbDialect(), err)
	}
}

// MigrateDb creates database tables and inserts default fixtures as needed.
func MigrateDb(dropDeprecated bool) {
	if dropDeprecated {
		DeprecatedTables.Drop()
	}

	Entities.Migrate()
	Entities.WaitForMigration()

	MigrateIndexes()

	CreateDefaultFixtures()
}

// ResetTestFixtures re-creates registered database tables and inserts test fixtures.
func ResetTestFixtures() {
	Entities.Migrate()
	Entities.WaitForMigration()
	Entities.Truncate()

	CreateDefaultFixtures()

	CreateTestFixtures()
}

// InitTestDb connects to and completely initializes the test database incl fixtures.
func InitTestDb(driver, dsn string) *Gorm {
	if HasDbProvider() {
		return nil
	}

	if driver == "test" || driver == "sqlite" || driver == "" || dsn == "" {
		driver = "sqlite3"
		dsn = ".test.db"
	}

	log.Infof("initializing %s test db in %s", driver, dsn)

	db := &Gorm{
		Driver: driver,
		Dsn:    dsn,
	}

	SetDbProvider(db)
	ResetTestFixtures()

	return db
}
