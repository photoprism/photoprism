package migrate

import (
	"sync"

	"github.com/photoprism/photoprism/pkg/dsn"
)

var Dialects = map[string]Migrations{
	dsn.DriverMySQL:   DialectMySQL,
	dsn.DriverSQLite3: DialectSQLite3,
}

var once = map[string]*sync.Once{
	dsn.DriverMySQL:   {},
	dsn.DriverSQLite3: {},
}
