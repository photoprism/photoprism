package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDriver(t *testing.T) {
	t.Run("MySQL", func(t *testing.T) {
		assert.Equal(t, DriverMySQL, ParseDriver("mysql"))
		assert.Equal(t, DriverMySQL, ParseDriver("MySQL"))
		assert.Equal(t, DriverMySQL, ParseDriver(" mysql "))
	})
	t.Run("MariaDB", func(t *testing.T) {
		// MariaDB and MySQL share a wire protocol and GORM dialect, so the
		// parser collapses both onto DriverMySQL.
		assert.Equal(t, DriverMySQL, ParseDriver("mariadb"))
		assert.Equal(t, DriverMySQL, ParseDriver("MariaDB"))
	})
	t.Run("Postgres", func(t *testing.T) {
		assert.Equal(t, DriverPostgres, ParseDriver("postgres"))
		// URI-style DSNs spell the driver "postgresql"; accept it as a Postgres alias.
		assert.Equal(t, DriverPostgres, ParseDriver("postgresql"))
		assert.Equal(t, DriverPostgres, ParseDriver("PostgreSQL"))
	})
	t.Run("SQLite3", func(t *testing.T) {
		assert.Equal(t, DriverSQLite3, ParseDriver("sqlite3"))
		assert.Equal(t, DriverSQLite3, ParseDriver("SQLite3"))
	})
	t.Run("SQLiteAliases", func(t *testing.T) {
		// Aliases that have historically resolved to the SQLite driver.
		assert.Equal(t, DriverSQLite3, ParseDriver("sqlite"))
		assert.Equal(t, DriverSQLite3, ParseDriver("test"))
		assert.Equal(t, DriverSQLite3, ParseDriver("file"))
	})
	t.Run("Empty", func(t *testing.T) {
		// Empty defaults to SQLite, matching the legacy config switch.
		assert.Equal(t, DriverSQLite3, ParseDriver(""))
		assert.Equal(t, DriverSQLite3, ParseDriver("   "))
	})
	t.Run("TiDB", func(t *testing.T) {
		// Recognized so callers can treat it as deprecated; not a supported driver.
		assert.Equal(t, DriverTiDB, ParseDriver("tidb"))
		assert.Equal(t, DriverTiDB, ParseDriver("TiDB"))
	})
	t.Run("Unknown", func(t *testing.T) {
		// Unknown inputs return an empty string so the caller's `default` arm fires.
		assert.Equal(t, "", ParseDriver("oracle"))
		assert.Equal(t, "", ParseDriver("garbage"))
	})
}
