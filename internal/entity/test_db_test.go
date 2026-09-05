package entity

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// testMysqlDSN returns the configured MySQL test DSN, or skips the test.
func testMysqlDSN(t *testing.T) string {
	t.Helper()

	if os.Getenv("PHOTOPRISM_TEST_DRIVER") != dsn.DriverMySQL {
		t.Skip("test database is not MySQL")
	}

	return os.Getenv("PHOTOPRISM_TEST_DSN")
}

func TestTestDbDSN(t *testing.T) {
	ValidateFixtures(t)
	t.Run("SQLite", func(t *testing.T) {
		assert.Equal(t, ".test.db", TestDbDSN(dsn.DriverSQLite3, ".test.db"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", TestDbDSN(dsn.DriverMySQL, ""))
	})
	t.Run("InvalidDSN", func(t *testing.T) {
		assert.Equal(t, "not a dsn", TestDbDSN(dsn.DriverMySQL, "not a dsn"))
	})
	t.Run("MySQL", func(t *testing.T) {
		base := testMysqlDSN(t)
		conf, err := mysql.ParseDSN(base)
		require.NoError(t, err)

		isolated, err := mysql.ParseDSN(TestDbDSN(dsn.DriverMySQL, base))
		require.NoError(t, err)

		assert.Equal(t, testDbName(conf.DBName), isolated.DBName)
		assert.NotEqual(t, conf.DBName, isolated.DBName)
		assert.Equal(t, conf.Addr, isolated.Addr)
		assert.Equal(t, conf.User, isolated.User)
	})
	t.Run("Cached", func(t *testing.T) {
		base := testMysqlDSN(t)
		assert.Equal(t, TestDbDSN(dsn.DriverMySQL, base), TestDbDSN(dsn.DriverMySQL, base))
	})
}

func TestTestDbName(t *testing.T) {
	ValidateFixtures(t)
	t.Run("PackageSuffix", func(t *testing.T) {
		name := testDbName("acceptance")
		assert.True(t, strings.HasPrefix(name, "acceptance_entity_"), name)
		assert.Equal(t, name, testDbName("acceptance"))
	})
	t.Run("MaxLen", func(t *testing.T) {
		name := testDbName(strings.Repeat("x", 100))
		assert.Equal(t, testDbNameLen, len(name))
		assert.True(t, strings.HasPrefix(name, "xxx"), name)
	})
	t.Run("InvalidChars", func(t *testing.T) {
		assert.Equal(t, testDbName("acceptance"), testDbName("Accept-ance!"))
	})
}

func TestCleanTestDbName(t *testing.T) {
	ValidateFixtures(t)
	t.Run("Lowercase", func(t *testing.T) {
		assert.Equal(t, "acceptance", cleanTestDbName("Acceptance"))
	})
	t.Run("RemovesInvalidChars", func(t *testing.T) {
		assert.Equal(t, "dropdatabase", cleanTestDbName("drop; database`"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", cleanTestDbName(""))
	})
}

func TestCreateTestDb(t *testing.T) {
	ValidateFixtures(t)
	t.Run("Success", func(t *testing.T) {
		conf, err := mysql.ParseDSN(testMysqlDSN(t))
		require.NoError(t, err)

		name := testDbName("createtestdb")

		serverConf := conf.Clone()
		serverConf.DBName = ""

		t.Cleanup(func() {
			db, openErr := sql.Open(dsn.DriverMySQL, serverConf.FormatDSN())
			require.NoError(t, openErr)
			defer db.Close()
			_, dropErr := db.Exec("DROP DATABASE IF EXISTS " + name)
			assert.NoError(t, dropErr)
		})

		assert.NoError(t, createTestDb(conf, name))
		// Creating it a second time must not fail.
		assert.NoError(t, createTestDb(conf, name))
	})
	t.Run("InvalidName", func(t *testing.T) {
		conf, err := mysql.ParseDSN(testMysqlDSN(t))
		require.NoError(t, err)

		assert.Error(t, createTestDb(conf, "invalid-name"))
	})
}
