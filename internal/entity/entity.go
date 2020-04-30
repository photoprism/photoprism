/*
Package entity contains models for data storage based on GORM.

See http://gorm.io/docs/ for more information about GORM.

Additional information concerning data storage can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki/Storage
*/
package entity

import (
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

func logError(result *gorm.DB) {
	if result.Error != nil {
		log.Error(result.Error.Error())
	}
}

// Migrate creates all tables and inserts default entities as needed.
func Migrate() {
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
func DropTables(db *gorm.DB) {
	db.DropTableIfExists(
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
