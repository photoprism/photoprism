package migrate

// Supported database dialects.
const (
	MySQL   = "mysql"
	SQLite3 = "sqlite3"
)

var Dialects = map[string]Migrations{
	MySQL:   DialectMySQL,
	SQLite3: DialectSQLite3,
}
