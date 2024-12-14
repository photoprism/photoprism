package migrate

import "sync"

// Supported database dialects.
const (
	MySQL    = "mysql"
	SQLite3  = "sqlite"
	Postgres = "postgres"
)

var Dialects = map[string]Migrations{
	MySQL:    DialectMySQL,
	SQLite3:  DialectSQLite,
	Postgres: DialectPostgres,
}

var once = map[string]*sync.Once{
	MySQL:    {},
	SQLite3:  {},
	Postgres: {},
}
