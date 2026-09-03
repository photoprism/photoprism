package dsn

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Gorm official dialect names
const (
	DialectMySQL      = "mysql"
	DialectPostgreSQL = "postgres"
	DialectSQLite     = "sqlite"
	DialectTiDB       = "tidb"
)

// DialectFromDriver canonicalizes a user-supplied driver identifier to one of the
// DialectMySQL/DialectPostgreSQL/DialectSQLite/DialectTiDB constants (case- and
// whitespace-insensitive).
func DialectFromDriver(s string) string {
	switch ParseDriver(s) {
	case DriverMySQL, DriverMariaDB:
		return DialectMySQL
	case DriverPostgres, DriverPostgreSQL:
		return DialectPostgreSQL
	case DriverSQLite3:
		return DialectSQLite
	case DriverTiDB:
		return DialectTiDB
	default:
		return ""
	}
}

// GormDrivers maps drivers to gorm dialectors
var GormDrivers = map[string]func(string) gorm.Dialector{
	DriverMySQL:      mysql.Open,
	DriverSQLite3:    sqlite.Open,
	DriverPostgres:   postgres.Open,
	DriverPostgreSQL: postgres.Open,
}
