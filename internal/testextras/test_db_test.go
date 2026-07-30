package testextras

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/photoprism/photoprism/pkg/dsn"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestDbDSN(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", TestDbDSN(dsn.DriverMySQL, ""))
	})
	t.Run("SQLite", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "sqlitefile")
		t.Setenv("PHOTOPRISM_TEST_DSN_SQLITEFILE", "file:/go/src/github.com/photoprism/photoprism/storage/testdata/unit.test.db?_foreign_keys=on&_busy_timeout=5000")
		sDSN := TestDbDSN(dsn.DriverSQLite3, "testdb")
		assert.NotEqual(t, sDSN, "file:/go/src/github.com/photoprism/photoprism/storage/testdata/testdb.db?_busy_timeout=5000&_foreign_keys=on")
		assert.True(t, strings.HasPrefix(sDSN, "file:/go/src/github.com/photoprism/photoprism/storage/testdata/testdb"))
		require.NoError(t, TestDbRemoveByName(dsn.DriverSQLite3, "testdb"))
	})
	t.Run("CachedSQLite", func(t *testing.T) {
		assert.Equal(t, TestDbDSN(dsn.DriverSQLite3, "testdb"), TestDbDSN(dsn.DriverSQLite3, "testdb"))
		require.NoError(t, TestDbRemoveByName(dsn.DriverSQLite3, "testdb"))
	})
	t.Run("MySQL", func(t *testing.T) {
		isolated := dsn.Parse(TestDbDSN(dsn.DriverMySQL, "testdb"))
		assert.NotEqual(t, testDbName("testdb"), isolated.Name)
		assert.NotEqual(t, "testdb", isolated.Name)
		assert.Equal(t, "mariadb:4001", isolated.Server)
		assert.Equal(t, "testdb", isolated.User)
		require.NoError(t, TestDbRemoveByName(dsn.DriverMySQL, "testdb"))
	})
	t.Run("CachedMySQL", func(t *testing.T) {
		assert.Equal(t, TestDbDSN(dsn.DriverMySQL, "testdb"), TestDbDSN(dsn.DriverMySQL, "testdb"))
		require.NoError(t, TestDbRemoveByName(dsn.DriverMySQL, "testdb"))
	})
	t.Run("Postgres", func(t *testing.T) {
		isolated := dsn.Parse(TestDbDSN(dsn.DriverPostgreSQL, "testdb"))
		assert.NotEqual(t, testDbName("testdb"), isolated.Name)
		assert.NotEqual(t, "testdb", isolated.Name)
		assert.Equal(t, "postgres:4002", isolated.Server)
		assert.Equal(t, "testdb", isolated.User)
		require.NoError(t, TestDbRemoveByDSN(dsn.DriverPostgreSQL, "testdb"))
	})
	t.Run("CachedPostgres", func(t *testing.T) {
		assert.Equal(t, TestDbDSN(dsn.DriverPostgreSQL, "testdb"), TestDbDSN(dsn.DriverPostgreSQL, "testdb"))
		require.NoError(t, TestDbRemoveByDSN(dsn.DriverPostgreSQL, "testdb"))
	})
}

func TestTestDbName(t *testing.T) {
	t.Run("PackageSuffix", func(t *testing.T) {
		name := testDbName("acceptance")
		assert.True(t, strings.HasPrefix(name, "acceptance_testextras_"), name)
		assert.NotEqual(t, name, testDbName("acceptance"))
	})
	t.Run("MaxLen", func(t *testing.T) {
		name := testDbName(strings.Repeat("x", 100))
		assert.Equal(t, testDbNameLen, len(name))
		assert.True(t, strings.HasPrefix(name, "xxx"), name)
	})
	t.Run("InvalidChars", func(t *testing.T) {
		name := testDbName("Accept-ance!")
		assert.True(t, strings.HasPrefix(name, "acceptance_testextras_"), name)
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
		_, dsname := dsn.PhotoPrismDriverToDriverDSN(dsn.DriverMySQL)
		conf := dsn.Parse(dsname)
		name := testDbName("migrate")

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
		_, dsname := dsn.PhotoPrismDriverToDriverDSN(dsn.DriverMySQL)
		conf := dsn.Parse(dsname)

		assert.Error(t, createTestDb(&conf, "invalid-name"))
	})
	t.Run("PostgresSuccess", func(t *testing.T) {
		_, dsname := dsn.PhotoPrismDriverToDriverDSN(dsn.DriverPostgres)
		conf := dsn.Parse(dsname)
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
		_, dsname := dsn.PhotoPrismDriverToDriverDSN(dsn.DriverPostgres)
		conf := dsn.Parse(dsname)

		assert.Error(t, createTestDb(&conf, "invalid-name"))
	})
}

func TestTestDbFromDSN(t *testing.T) {
	t.Run("MySQLSuccess", func(t *testing.T) {
		mdbDSN := dsn.Parse(TestDbFromDSN(dsn.TestDSNFromEnv(dsn.DriverMariaDB, "migrate")))
		assert.NotEqual(t, "migrate", mdbDSN.Name)
		require.NoError(t, TestDbRemoveByName(mdbDSN.Driver, "migrate"))
	})
	t.Run("PostgresSuccess", func(t *testing.T) {
		pdbDSN := dsn.Parse(TestDbFromDSN(dsn.TestDSNFromEnv(dsn.DriverPostgreSQL, "migrate")))
		assert.NotEqual(t, "migrate", pdbDSN.Name)
		require.NoError(t, TestDbRemoveByName(pdbDSN.Driver, "migrate"))
	})
	t.Run("BothSuccess", func(t *testing.T) {
		mdbDSN := dsn.Parse(TestDbFromDSN(dsn.TestDSNFromEnv(dsn.DriverMariaDB, "migrate")))
		pdbDSN := dsn.Parse(TestDbFromDSN(dsn.TestDSNFromEnv(dsn.DriverPostgreSQL, "migrate")))

		assert.NotEqual(t, "migrate", mdbDSN.Name)
		assert.NotEqual(t, "migrate", pdbDSN.Name)
		assert.NotEqual(t, mdbDSN.Name, pdbDSN.Name)
		assert.NotEqual(t, mdbDSN.ToString(), pdbDSN.ToString())
		pdbDSN2 := dsn.Parse(TestDbFromDSN(dsn.TestDSNFromEnv(dsn.DriverPostgreSQL, "migrate")))
		assert.Equal(t, pdbDSN2.Name, pdbDSN.Name)

		require.NoError(t, TestDbRemoveByName(mdbDSN.Driver, "migrate"))
		require.NoError(t, TestDbRemoveByName(pdbDSN.Driver, "migrate"))
	})
	t.Run("SQLite", func(t *testing.T) {
		sdbDSN := dsn.Parse(TestDbFromDSN(dsn.TestDSNFromEnv(dsn.DriverSQLite3, "migrate")))
		assert.NotEqual(t, "migrate.db", sdbDSN.Name)
		assert.NotEqual(t, "migrate", sdbDSN.Name)
		sdbDSN2 := dsn.Parse(TestDbFromDSN(dsn.TestDSNFromEnv(dsn.DriverSQLite3, "migrate")))
		assert.Equal(t, sdbDSN2.Name, sdbDSN.Name)
		require.NoError(t, TestDbRemoveByName(sdbDSN.Driver, "migrate"))
	})
}
