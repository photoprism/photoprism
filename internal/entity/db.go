package entity

import "github.com/jinzhu/gorm"

// Db returns the default *gorm.DB connection.
func Db() *gorm.DB {
	return dbConn.Db()
}

// UnscopedDb returns an unscoped *gorm.DB connection
// that returns all records including deleted records.
func UnscopedDb() *gorm.DB {
	return Db().Unscoped()
}
