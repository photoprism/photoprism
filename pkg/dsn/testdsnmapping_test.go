package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhotoPrismTestToDriverDSN(t *testing.T) {
	t.Run("sqlite", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "sqlite")
		t.Setenv("PHOTOPRISM_TEST_DSN_SQLITE", ":memory:?cache=shared&_foreign_keys=on")

		driver, dsn := PhotoPrismTestToDriverDSN(0)

		assert.Equal(t, "sqlite", driver)
		assert.Equal(t, ":memory:?cache=shared&_foreign_keys=on", dsn)
		driver, dsn = PhotoPrismTestToDriverDSN(1)

		assert.Equal(t, "sqlite", driver)
		assert.Equal(t, ":memory:?cache=shared&_foreign_keys=on", dsn)
	})

	t.Run("sqlitefile", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "sqlitefile")
		t.Setenv("PHOTOPRISM_TEST_DSN_SQLITEFILE", "file:/go/src/github.com/photoprism/photoprism/storage/testdata/unit.test.db?_foreign_keys=on&_busy_timeout=5000")

		driver, dsn := PhotoPrismTestToDriverDSN(0)

		assert.Equal(t, "sqlite", driver)
		assert.Equal(t, "file:/go/src/github.com/photoprism/photoprism/storage/testdata/unit.test.db?_foreign_keys=on&_busy_timeout=5000", dsn)
		driver, dsn = PhotoPrismTestToDriverDSN(1)

		assert.Equal(t, "sqlite", driver)
		assert.Equal(t, "file:/go/src/github.com/photoprism/photoprism/storage/testdata/unit.test.db_01.db?_foreign_keys=on&_busy_timeout=5000", dsn)
	})

	t.Run("mariadb", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "mariadb")
		t.Setenv("PHOTOPRISM_TEST_DSN_MARIADB", "root:photoprism@tcp(mariadb:4001)/testdb?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true")

		driver, dsn := PhotoPrismTestToDriverDSN(0)

		assert.Equal(t, "mysql", driver)
		assert.Equal(t, "root:photoprism@tcp(mariadb:4001)/testdb?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true", dsn)
		driver, dsn = PhotoPrismTestToDriverDSN(1)

		assert.Equal(t, "mysql", driver)
		assert.Equal(t, "root:photoprism@tcp(mariadb:4001)/testdb_01?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true", dsn)
	})

	t.Run("mysql8", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "mysql8")
		t.Setenv("PHOTOPRISM_TEST_DSN_MYSQL8", "root:photoprism@tcp(mysql:4001)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=15s")

		driver, dsn := PhotoPrismTestToDriverDSN(0)

		assert.Equal(t, "mysql", driver)
		assert.Equal(t, "root:photoprism@tcp(mysql:4001)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=15s", dsn)
		driver, dsn = PhotoPrismTestToDriverDSN(1)

		assert.Equal(t, "mysql", driver)
		assert.Equal(t, "root:photoprism@tcp(mysql:4001)/photoprism_01?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=15s", dsn)
	})

	t.Run("postgres", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "postgres")
		t.Setenv("PHOTOPRISM_TEST_DSN_POSTGRES", "postgresql://testdb:testdb@postgres:5432/testdb?TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable")

		driver, dsn := PhotoPrismTestToDriverDSN(0)

		assert.Equal(t, "postgres", driver)
		assert.Equal(t, "postgresql://testdb:testdb@postgres:5432/testdb?TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable", dsn)
		driver, dsn = PhotoPrismTestToDriverDSN(1)

		assert.Equal(t, "postgres", driver)
		assert.Equal(t, "postgresql://testdb:testdb@postgres:5432/testdb_01?TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable", dsn)
	})

	t.Run("default", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "unknown")

		driver, dsn := PhotoPrismTestToDriverDSN(0)

		assert.Equal(t, "sqlite", driver)
		assert.Equal(t, "", dsn)
	})
}
