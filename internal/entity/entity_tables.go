package entity

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/migrate"
	"github.com/photoprism/photoprism/pkg/clean"
)

type Tables map[string]interface{}

// Entities contains database entities and their table names.
var Entities = Tables{
	migrate.Migration{}.TableName(): &migrate.Migration{},
	migrate.Version{}.TableName():   &migrate.Version{},
	Error{}.TableName():             &Error{},
	Password{}.TableName():          &Password{},
	User{}.TableName():              &User{},
	UserDetails{}.TableName():       &UserDetails{},
	UserSettings{}.TableName():      &UserSettings{},
	Session{}.TableName():           &Session{},
	Client{}.TableName():            &Client{},
	Service{}.TableName():           &Service{},
	Folder{}.TableName():            &Folder{},
	Duplicate{}.TableName():         &Duplicate{},
	File{}.TableName():              &File{},
	FileShare{}.TableName():         &FileShare{},
	FileSync{}.TableName():          &FileSync{},
	Photo{}.TableName():             &Photo{},
	PhotoUser{}.TableName():         &PhotoUser{},
	Details{}.TableName():           &Details{},
	Place{}.TableName():             &Place{},
	Cell{}.TableName():              &Cell{},
	Camera{}.TableName():            &Camera{},
	Lens{}.TableName():              &Lens{},
	Country{}.TableName():           &Country{},
	Album{}.TableName():             &Album{},
	AlbumUser{}.TableName():         &AlbumUser{},
	PhotoAlbum{}.TableName():        &PhotoAlbum{},
	Label{}.TableName():             &Label{},
	Category{}.TableName():          &Category{},
	PhotoLabel{}.TableName():        &PhotoLabel{},
	Keyword{}.TableName():           &Keyword{},
	PhotoKeyword{}.TableName():      &PhotoKeyword{},
	Link{}.TableName():              &Link{},
	Subject{}.TableName():           &Subject{},
	Face{}.TableName():              &Face{},
	Marker{}.TableName():            &Marker{},
	Reaction{}.TableName():          &Reaction{},
	UserShare{}.TableName():         &UserShare{},
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
				log.Tracef("migrate: %s migrated", clean.Log(name))
				break
			} else {
				log.Tracef("migrate: waiting for %s migration (%s)", clean.Log(name), err.Error())
				time.Sleep(100 * time.Millisecond)
			}

			if i == attempts {
				panic("migration failed")
			}
		}
	}
}

// Truncate removes all data from tables without dropping them.
func (list Tables) Truncate(db *gorm.DB) {
	var name string

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("migrate: %s in %s (truncate)", r, name)
		}
	}()

	for name = range list {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE 1", name)).Error; err == nil {
			// log.Debugf("entity: removed all data from %s", name)
			break
		} else if err.Error() != "record not found" {
			log.Debugf("migrate: %s in %s", err, clean.Log(name))
		}
	}
}

// Migrate migrates all database tables of registered entities.
func (list Tables) Migrate(db *gorm.DB, opt migrate.Options) {
	var name string
	var entity interface{}

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("migrate: %s in %s (panic)", r, name)
		}
	}()

	log.Debugf("migrate: running database migrations")

	// Run pre migrations, if any.
	if err := migrate.Run(db, opt.Pre()); err != nil {
		log.Error(err)
	}

	// Run ORM auto migrations.
	if opt.AutoMigrate {
		for name, entity = range list {
			if err := db.AutoMigrate(entity).Error; err != nil {
				log.Debugf("migrate: %s (waiting 1s)", err.Error())

				time.Sleep(time.Second)

				if err = db.AutoMigrate(entity).Error; err != nil {
					log.Errorf("migrate: failed migrating %s", clean.Log(name))
					panic(err)
				}
			}
		}
	}

	// Run main migrations, if any.
	if err := migrate.Run(db, opt); err != nil {
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
