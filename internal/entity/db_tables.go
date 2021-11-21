package entity

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/migrate"
	"github.com/photoprism/photoprism/pkg/txt"
)

type Tables map[string]interface{}

// Entities contains database entities and their table names.
var Entities = Tables{
	migrate.Migration{}.TableName(): &migrate.Migration{},
	"errors":                        &Error{},
	"addresses":                     &Address{},
	"users":                         &User{},
	"accounts":                      &Account{},
	"folders":                       &Folder{},
	"duplicates":                    &Duplicate{},
	File{}.TableName():              &File{},
	"files_share":                   &FileShare{},
	"files_sync":                    &FileSync{},
	Photo{}.TableName():             &Photo{},
	"details":                       &Details{},
	Place{}.TableName():             &Place{},
	Cell{}.TableName():              &Cell{},
	"cameras":                       &Camera{},
	"lenses":                        &Lens{},
	"countries":                     &Country{},
	"albums":                        &Album{},
	"photos_albums":                 &PhotoAlbum{},
	"labels":                        &Label{},
	"categories":                    &Category{},
	"photos_labels":                 &PhotoLabel{},
	"keywords":                      &Keyword{},
	"photos_keywords":               &PhotoKeyword{},
	"passwords":                     &Password{},
	"links":                         &Link{},
	Subject{}.TableName():           &Subject{},
	Face{}.TableName():              &Face{},
	Marker{}.TableName():            &Marker{},
}

// WaitForMigration waits for the database migration to be successful.
func (list Tables) WaitForMigration() {
	type RowCount struct {
		Count int
	}

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
func (list Tables) Truncate() {
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
func (list Tables) Migrate() {
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

	if err := migrate.Auto(Db()); err != nil {
		log.Error(err)
	}
}

// Drop drops all database tables of registered entities.
func (list Tables) Drop() {
	for _, entity := range list {
		if err := UnscopedDb().DropTableIfExists(entity).Error; err != nil {
			panic(err)
		}
	}
}
