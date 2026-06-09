package dsn

import "strings"

// SQL database drivers.
const (
	DriverNone     = ""         // Unknown or unsupported database driver.
	DriverAuto     = "auto"     // Automatic database driver selection.
	DriverMySQL    = "mysql"    // GORM dialect for MySQL/MariaDB; the canonical driver name PhotoPrism stores.
	DriverMariaDB  = "mariadb"  // Accepted as user input; ParseDriver collapses it to DriverMySQL since the dialect is shared.
	DriverPostgres = "postgres" // Reserved identifier; PostgreSQL is not a supported runtime target yet (requires a GORM upgrade).
	DriverSQLite3  = "sqlite3"  // GORM dialect for SQLite; the default when no driver is configured.
	DriverTiDB     = "tidb"     // Deprecated; recognized so callers can warn and fall back to SQLite.
)

// SQLite default DSNs.
const (
	SQLiteTestDB       = ".test.db"              // Default on-disk DSN for tests that need a fresh SQLite database.
	SQLiteMemory       = ":memory:"              // Bare in-memory DSN; each connection gets a separate database (rarely what tests want).
	SQLiteMemoryShared = ":memory:?cache=shared" // In-memory DSN with shared page cache; multiple connections share one database.
)

// ParseDriver canonicalizes a user-supplied driver identifier.
func ParseDriver(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case DriverMySQL, DriverMariaDB:
		return DriverMySQL
	case DriverPostgres, "postgresql":
		return DriverPostgres
	case DriverSQLite3, "sqlite", "test", "file":
		return DriverSQLite3
	case DriverTiDB:
		return DriverTiDB
	case DriverAuto:
		return DriverAuto
	default:
		return DriverNone
	}
}

// Params maps required DSN parameters by driver type.
var Params = Values{
	DriverMySQL:    "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
	DriverMariaDB:  "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
	DriverPostgres: "sslmode=disable TimeZone=UTC",
	DriverSQLite3:  "_busy_timeout=5000",
}
