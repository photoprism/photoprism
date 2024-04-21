package entity

import (
	"gorm.io/gorm"
)

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
