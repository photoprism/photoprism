package entity

import (
	"time"

	"gorm.io/gorm"
)

// Set UTC as the default for created and updated timestamps.
func init() {
	gormConfig().NowFunc = func() time.Time {
		return UTC()
	}
}

// Db returns the default *gorm.DB connection.
func Db() *gorm.DB {
	if dbConn == nil {
		return nil
	}

	return dbConn.Db()
}

// UnscopedDb returns an unscoped *gorm.DB connection
// that returns all records including deleted records.
func UnscopedDb() *gorm.DB {
	return Db().Unscoped()
}
