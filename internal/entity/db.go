package entity

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Set UTC as the default for created and updated timestamps.
func init() {
	gorm.NowFunc = func() time.Time {
		return UTC()
	}
}

// Db returns the default *gorm.DB connection.
func Db() *gorm.DB {
	return dbConn.Db()
}

// UnscopedDb returns an unscoped *gorm.DB connection
// that returns all records including deleted records.
func UnscopedDb() *gorm.DB {
	return Db().Unscoped()
}
