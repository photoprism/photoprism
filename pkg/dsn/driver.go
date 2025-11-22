package dsn

// SQL database drivers.
const (
	DriverMySQL    = "mysql"
	DriverMariaDB  = "mariadb"
	DriverPostgres = "postgres"
	DriverSQLite3  = "sqlite3"
)

// SQLite default DSNs.
const (
	SQLiteTestDB = ".test.db"
	SQLiteMemory = ":memory:"
)

// Params maps required DSN parameters by driver type.
var Params = Values{
	DriverMySQL:    "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
	DriverMariaDB:  "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
	DriverPostgres: "sslmode=disable TimeZone=UTC",
	DriverSQLite3:  "_busy_timeout=5000",
}
