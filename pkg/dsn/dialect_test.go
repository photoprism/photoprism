package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDialectFromDriver(t *testing.T) {
	t.Run("MySQL", func(t *testing.T) {
		assert.Equal(t, DialectMySQL, DialectFromDriver(DriverMySQL))
		assert.Equal(t, DialectMySQL, DialectFromDriver("MariaDB"))
	})
	t.Run("SQLite", func(t *testing.T) {
		assert.Equal(t, DialectSQLite, DialectFromDriver(DriverSQLite3))
		assert.Equal(t, DialectSQLite, DialectFromDriver(DialectSQLite))
		assert.Equal(t, DialectSQLite, DialectFromDriver("test"))
		assert.Equal(t, DialectSQLite, DialectFromDriver("file"))
	})
	t.Run("PostgreSQL", func(t *testing.T) {
		assert.Equal(t, DialectPostgreSQL, DialectFromDriver(DriverPostgres))
		assert.Equal(t, DialectPostgreSQL, DialectFromDriver("postgresql"))
	})
	t.Run("Unsupported", func(t *testing.T) {
		assert.Equal(t, "", DialectFromDriver(""))
		assert.Equal(t, "", DialectFromDriver(DriverTiDB))
		assert.Equal(t, "", DialectFromDriver("oracle"))
	})
}
