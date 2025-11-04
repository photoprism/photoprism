package dsn

// SQL database drivers.
const (
	DriverMySQL      = "mysql"
	DriverMariaDB    = "mariadb"
	DriverPostgres   = "postgres"
	DriverPostgreSQL = "postgresql"
	DriverSQLite3    = "sqlite"
)

// SQLite default DSNs.
const (
	SQLiteTestDB = ".test.db"
	SQLiteMemory = ":memory:?cache=shared&_foreign_keys=on"
)

// Params maps required DSN parameters by driver type.
var Params = Values{
	DriverMySQL:      "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
	DriverMariaDB:    "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
	DriverPostgres:   "sslmode=disable TimeZone=UTC lock_timeout=5000",
	DriverPostgreSQL: "sslmode=disable&TimeZone=UTC&lock_timeout=5000",
	DriverSQLite3:    "_busy_timeout=5000&_foreign_keys=on",
}
