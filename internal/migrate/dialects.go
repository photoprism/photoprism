package migrate

// Supported database dialects.
const (
	MySQL  = "mysql"
	SQLite = "sqlite3"
)

var Dialects = map[string]Migrations{
	MySQL:  DialectMySQL,
	SQLite: DialectSQLite,
}
