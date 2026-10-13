package entity

import (
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
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

// testPostgresDSN returns the configured PostgreSQL test DSN, or skips the test.
func testPostgresDSN(t *testing.T) string {
	t.Helper()

	if dsn.ParseDriver(os.Getenv("PHOTOPRISM_TEST_DRIVER")) != dsn.DriverPostgres {
		t.Skip("test database is not PostgreSQL")
	}

	return os.Getenv("PHOTOPRISM_TEST_DSN")
}

// openPostgresServer opens a connection to the database of a PostgreSQL DSN and closes it when the test ends.
func openPostgresServer(t *testing.T, serverDsn string) *sql.DB {
	t.Helper()

	config, err := pgx.ParseConfig(serverDsn)
	require.NoError(t, err)
	db := stdlib.OpenDB(*config)
	t.Cleanup(func() { _ = db.Close() })

	return db
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
		base := testMysqlDSN(t)
		conf, err := mysql.ParseDSN(base)
		require.NoError(t, err)

		isolated, err := mysql.ParseDSN(TestDbDSN(dsn.DriverMySQL, base))
		require.NoError(t, err)

		assert.Equal(t, testDbName(conf.DBName, testDbNameLen), isolated.DBName)
		assert.NotEqual(t, conf.DBName, isolated.DBName)
		assert.Equal(t, conf.Addr, isolated.Addr)
		assert.Equal(t, conf.User, isolated.User)
	})
	t.Run("Cached", func(t *testing.T) {
		base := testMysqlDSN(t)
		assert.Equal(t, TestDbDSN(dsn.DriverMySQL, base), TestDbDSN(dsn.DriverMySQL, base))
	})
	t.Run("PostgresKeyValue", func(t *testing.T) {
		s := "host=localhost user=u password=p dbname=acceptance"
		assert.Equal(t, s, TestDbDSN(dsn.DriverPostgres, s))
	})
	t.Run("PostgreSQL", func(t *testing.T) {
		base := testPostgresDSN(t)
		baseUri, err := url.Parse(base)
		require.NoError(t, err)

		isolated, err := url.Parse(TestDbDSN(dsn.DriverPostgres, base))
		require.NoError(t, err)

		assert.Equal(t, "/"+testDbName(strings.TrimPrefix(baseUri.Path, "/"), testPostgresDbNameLen), isolated.Path)
		assert.NotEqual(t, baseUri.Path, isolated.Path)
		assert.Equal(t, baseUri.Host, isolated.Host)
		assert.Equal(t, baseUri.RawQuery, isolated.RawQuery)
		assert.Equal(t, TestDbDSN("postgresql", base), isolated.String())
	})
}

func TestTestDbName(t *testing.T) {
	t.Run("PackageSuffix", func(t *testing.T) {
		name := testDbName("testdb", testDbNameLen)
		assert.True(t, strings.HasPrefix(name, "testdb_entity_"), name)
		assert.Equal(t, name, testDbName("testdb", testDbNameLen))
	})
	t.Run("MaxLen", func(t *testing.T) {
		name := testDbName(strings.Repeat("x", 100), testDbNameLen)
		assert.Equal(t, testDbNameLen, len(name))
		assert.True(t, strings.HasPrefix(name, "xxx"), name)
		assert.Len(t, testDbName(strings.Repeat("x", 100), testPostgresDbNameLen), testPostgresDbNameLen)
	})
	t.Run("InvalidChars", func(t *testing.T) {
		assert.Equal(t, testDbName("acceptance", testDbNameLen), testDbName("Accept-ance!", testDbNameLen))
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
	t.Run("Success", func(t *testing.T) {
		conf, err := mysql.ParseDSN(testMysqlDSN(t))
		require.NoError(t, err)

		name := testDbName("createtestdb", testDbNameLen)

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

func TestCreatePostgresTestDb(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		base := testPostgresDSN(t)
		name := testDbName("createtestdb", testPostgresDbNameLen)
		server := openPostgresServer(t, base)

		t.Cleanup(func() {
			_, dropErr := server.Exec("DROP DATABASE IF EXISTS " + name)
			assert.NoError(t, dropErr)
		})

		assert.NoError(t, createPostgresTestDb(base, name))
		// Creating it a second time must not fail.
		assert.NoError(t, createPostgresTestDb(base, name))

		var provider string
		require.NoError(t, server.QueryRow("SELECT datlocprovider FROM pg_database WHERE datname = $1", name).Scan(&provider))
		assert.Equal(t, "i", provider)
	})
	t.Run("MissingDatabase", func(t *testing.T) {
		base, err := url.Parse(testPostgresDSN(t))
		require.NoError(t, err)
		name := testDbName("createtestdbmissing", testPostgresDbNameLen)
		server := openPostgresServer(t, base.String())
		t.Cleanup(func() { _, _ = server.Exec("DROP DATABASE IF EXISTS " + name) })

		// The database of the DSN does not exist yet, as when a test names a new one.
		base.Path = "/" + name
		assert.NoError(t, createPostgresTestDb(base.String(), name))

		var exists bool
		require.NoError(t, server.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists))
		assert.True(t, exists)
	})
	t.Run("InvalidName", func(t *testing.T) {
		assert.Error(t, createPostgresTestDb("postgres://u:p@localhost/db", "invalid-name"))
	})
	t.Run("InvalidDSN", func(t *testing.T) {
		assert.Error(t, createPostgresTestDb("postgres://u:p@localhost:bad/db", "valid_name"))
	})
}

func TestTestMySQLDbDSN(t *testing.T) {
	t.Run("InvalidDSN", func(t *testing.T) {
		_, err := testMySQLDbDSN("not a dsn")
		assert.Error(t, err)
	})
	t.Run("Success", func(t *testing.T) {
		base := testMysqlDSN(t)
		isolated, err := testMySQLDbDSN(base)
		require.NoError(t, err)
		assert.NotEqual(t, base, isolated)
	})
}

func TestTestPostgresDbDSN(t *testing.T) {
	t.Run("NotURI", func(t *testing.T) {
		_, err := testPostgresDbDSN("host=localhost dbname=acceptance")
		assert.Error(t, err)
	})
}
