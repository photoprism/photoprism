package migrate

import (
	"sync"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// Dialects maps a database driver to the migrations written for it.
var Dialects = map[string]Migrations{
	dsn.DriverMySQL:      DialectMySQL,
	dsn.DriverSQLite3:    DialectSQLite,
	dsn.DriverPostgres:   DialectPostgres,
	dsn.DriverPostgreSQL: DialectPostgres,
}

var once = map[string]*sync.Once{
	dsn.DriverMySQL:      {},
	dsn.DriverSQLite3:    {},
	dsn.DriverPostgres:   {},
	dsn.DriverPostgreSQL: {},
}
