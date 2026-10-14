package dsn

// SQL dialects, named as the ORM drivers name them. A dialect identifies the SQL a database speaks,
// while a driver is the configured name, e.g. "sqlite3" selects the SQLite dialect.
const (
	DialectMySQL      = "mysql"
	DialectPostgreSQL = "postgres"
	DialectSQLite     = "sqlite"
)

// DialectFromDriver returns the SQL dialect of a driver or dialect name, or an empty string if it
// is not supported.
func DialectFromDriver(s string) string {
	switch ParseDriver(s) {
	case DriverMySQL:
		return DialectMySQL
	case DriverPostgres:
		return DialectPostgreSQL
	case DriverSQLite3:
		return DialectSQLite
	default:
		return ""
	}
}
