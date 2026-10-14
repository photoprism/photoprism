package migrate

import (
	"sync"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// DialectPostgreSQL lists the migrations that run on PostgreSQL, whose schema is created by the ORM alone.
var DialectPostgreSQL = Migrations{}

// Dialects maps a database driver to the migrations written for it.
var Dialects = map[string]Migrations{
	dsn.DriverMySQL:    DialectMySQL,
	dsn.DriverPostgres: DialectPostgreSQL,
	dsn.DriverSQLite3:  DialectSQLite3,
}

var once = map[string]*sync.Once{
	dsn.DriverMySQL:    {},
	dsn.DriverPostgres: {},
	dsn.DriverSQLite3:  {},
}
