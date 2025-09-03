package constants

// Supported database(s)
const (
	MySQL   = "mysql"
	MariaDB = "mariadb"
	SQLite3 = "sqlite3"
)

// Future database(s)
const (
	Postgres = "postgres"
)

// Test database(s)
const (
	SQLiteTestDB    = ".test.db"
	SQLiteMemoryDSN = ":memory:?cache=shared"
)
