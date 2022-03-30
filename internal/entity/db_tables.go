package entity

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/migrate"
	"github.com/photoprism/photoprism/pkg/sanitize"
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
func (list Tables) WaitForMigration(db *gorm.DB) {
	type RowCount struct {
		Count int
	}

	attempts := 100
	for name := range list {
		for i := 0; i <= attempts; i++ {
			count := RowCount{}
			if err := db.Raw(fmt.Sprintf("SELECT COUNT(*) AS count FROM %s", name)).Scan(&count).Error; err == nil {
				log.Tracef("entity: %s migrated", sanitize.Log(name))
				break
			} else {
				log.Debugf("entity: waiting for %s migration (%s)", sanitize.Log(name), err.Error())
			}

			if i == attempts {
				panic("migration failed")
			}

			time.Sleep(50 * time.Millisecond)
		}
	}
}

// Truncate removes all data from tables without dropping them.
func (list Tables) Truncate(db *gorm.DB) {
	for name := range list {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE 1", name)).Error; err == nil {
			// log.Debugf("entity: removed all data from %s", name)
			break
		} else if err.Error() != "record not found" {
			log.Debugf("entity: %s in %s", err, sanitize.Log(name))
		}
	}
}

// Migrate migrates all database tables of registered entities.
func (list Tables) Migrate(db *gorm.DB, runFailed bool, ids []string) {
	if len(ids) == 0 {
		for name, entity := range list {
			if err := db.AutoMigrate(entity).Error; err != nil {
				log.Debugf("entity: %s (waiting 1s)", err.Error())

				time.Sleep(time.Second)

				if err := db.AutoMigrate(entity).Error; err != nil {
					log.Errorf("entity: failed migrating %s", sanitize.Log(name))
					panic(err)
				}
			}
		}
	}

	if err := migrate.Auto(db, runFailed, ids); err != nil {
		log.Error(err)
	}
}

// Drop drops all database tables of registered entities.
func (list Tables) Drop(db *gorm.DB) {
	for _, entity := range list {
		if err := db.DropTableIfExists(entity).Error; err != nil {
			panic(err)
		}
	}
}
