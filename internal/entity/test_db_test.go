package entity

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// testMariaDBDSN returns the configured MySQL test DSN, or skips the test.
func testMariaDBDSN(t *testing.T) string {
	t.Helper()

	if os.Getenv("PHOTOPRISM_TEST_DSN_NAME") != dsn.DriverMariaDB {
		t.Skip("test database is not MariaDB")
	}

	return os.Getenv("PHOTOPRISM_TEST_DSN_MARIADB")
}

// testPostgresDSN returns the configured Postgres test DSN, or skips the test.
func testPostgresDSN(t *testing.T) string {
	t.Helper()

	if os.Getenv("PHOTOPRISM_TEST_DSN_NAME") != dsn.DriverPostgres {
		t.Skip("test database is not Postgres")
	}

	return os.Getenv("PHOTOPRISM_TEST_DSN_POSTGRES")
}

func TestTestDbDSN(t *testing.T) {
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
		base := testMariaDBDSN(t)
		conf := dsn.Parse(base)

		isolated := dsn.Parse(TestDbDSN(dsn.DriverMySQL, base))

		assert.Equal(t, testDbName(conf.Name), isolated.Name)
		assert.NotEqual(t, conf.Name, isolated.Name)
		assert.Equal(t, conf.Server, isolated.Server)
		assert.Equal(t, conf.User, isolated.User)
	})
	t.Run("CachedMySQL", func(t *testing.T) {
		base := testMariaDBDSN(t)
		assert.Equal(t, TestDbDSN(dsn.DriverMySQL, base), TestDbDSN(dsn.DriverMySQL, base))
	})
	t.Run("Postgres", func(t *testing.T) {
		base := testPostgresDSN(t)
		conf := dsn.Parse(base)

		isolated := dsn.Parse(TestDbDSN(dsn.DriverPostgreSQL, base))

		assert.Equal(t, testDbName(conf.Name), isolated.Name)
		assert.NotEqual(t, conf.Name, isolated.Name)
		assert.Equal(t, conf.Server, isolated.Server)
		assert.Equal(t, conf.User, isolated.User)
	})
	t.Run("CachedPostgres", func(t *testing.T) {
		base := testPostgresDSN(t)
		assert.Equal(t, TestDbDSN(dsn.DriverPostgreSQL, base), TestDbDSN(dsn.DriverPostgreSQL, base))
	})
}

func TestTestDbName(t *testing.T) {
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
	t.Run("MySQLSuccess", func(t *testing.T) {
		conf := dsn.Parse(testMariaDBDSN(t))

		name := testDbName("createtestdb")

		t.Cleanup(func() {
			db, openErr := sql.Open(dsn.DriverMySQL, conf.ToString())
			require.NoError(t, openErr)
			defer db.Close()
			_, dropErr := db.Exec("DROP DATABASE IF EXISTS " + name)
			assert.NoError(t, dropErr)
		})

		assert.NoError(t, createTestDb(&conf, name))
		// Creating it a second time must not fail.
		assert.NoError(t, createTestDb(&conf, name))
	})
	t.Run("MySQLInvalidName", func(t *testing.T) {
		conf := dsn.Parse(testMariaDBDSN(t))

		assert.Error(t, createTestDb(&conf, "invalid-name"))
	})
	t.Run("PostgresSuccess", func(t *testing.T) {
		conf := dsn.Parse(testPostgresDSN(t))

		name := testDbName("createtestdb")

		t.Cleanup(func() {
			db, openErr := sql.Open("pgx", conf.ToString())
			require.NoError(t, openErr)
			defer db.Close()
			_, dropErr := db.Exec("DROP DATABASE IF EXISTS " + name)
			assert.NoError(t, dropErr)
		})

		assert.NoError(t, createTestDb(&conf, name))
		// Creating it a second time must not fail.
		assert.NoError(t, createTestDb(&conf, name))
	})
	t.Run("PostgresInvalidName", func(t *testing.T) {
		conf := dsn.Parse(testPostgresDSN(t))

		assert.Error(t, createTestDb(&conf, "invalid-name"))
	})
}
