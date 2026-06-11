package entity

import (
	"github.com/jinzhu/gorm"
)

// Set UTC, truncated to whole seconds, as the default for created and updated
// timestamps. The schema stores these as DATETIME without fractional seconds, so
// generating second-precision values keeps in-memory and persisted times in sync
// and makes timestamp comparisons behave the same on SQLite and MariaDB.
func init() {
	gorm.NowFunc = Now
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
