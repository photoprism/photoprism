package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhotoPrismTestToDriverDSN(t *testing.T) {
	t.Run("sqlite", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "sqlite")
		t.Setenv("PHOTOPRISM_TEST_DSN_SQLITE", ":memory:?cache=shared&_foreign_keys=on")
		t.Setenv("PHOTOPRISM_TEST_DSN_SQLITEFILE", "file:/go/src/github.com/photoprism/photoprism/storage/testdata/unit.test.db?_foreign_keys=on&_busy_timeout=5000")

		driver, dsn := PhotoPrismTestToDriverDSN()

		assert.Equal(t, "sqlite", driver)
		assert.Equal(t, "file:/go/src/github.com/photoprism/photoprism/storage/testdata/unit.test.db?_foreign_keys=on&_busy_timeout=5000", dsn)
	})

	t.Run("sqlitefile", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "sqlitefile")
		t.Setenv("PHOTOPRISM_TEST_DSN_SQLITEFILE", "file:/go/src/github.com/photoprism/photoprism/storage/testdata/unit.test.db?_foreign_keys=on&_busy_timeout=5000")

		driver, dsn := PhotoPrismTestToDriverDSN()

		assert.Equal(t, "sqlite", driver)
		assert.Equal(t, "file:/go/src/github.com/photoprism/photoprism/storage/testdata/unit.test.db?_foreign_keys=on&_busy_timeout=5000", dsn)
	})

	t.Run("mariadb", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "mariadb")
		t.Setenv("PHOTOPRISM_TEST_DSN_MARIADB", "root:photoprism@tcp(mariadb:4001)/mysql?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true")

		driver, dsn := PhotoPrismTestToDriverDSN()

		assert.Equal(t, "mysql", driver)
		assert.Equal(t, "root:photoprism@tcp(mariadb:4001)/mysql?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true", dsn)
	})

	t.Run("mysql8", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "mysql8")
		t.Setenv("PHOTOPRISM_TEST_DSN_MYSQL8", "root:photoprism@tcp(mysql:4001)/mysql?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=15s")

		driver, dsn := PhotoPrismTestToDriverDSN()

		assert.Equal(t, "mysql", driver)
		assert.Equal(t, "root:photoprism@tcp(mysql:4001)/mysql?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=15s", dsn)
	})

	t.Run("postgres", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "postgres")
		t.Setenv("PHOTOPRISM_TEST_DSN_POSTGRES", "postgresql://photoprism:photoprism@postgres:5432/postgres?TimeZone=UTC&connect_timeout=15&sslmode=disable")

		driver, dsn := PhotoPrismTestToDriverDSN()

		assert.Equal(t, "postgres", driver)
		assert.Equal(t, "postgresql://photoprism:photoprism@postgres:5432/postgres?TimeZone=UTC&connect_timeout=15&sslmode=disable", dsn)
	})

	t.Run("default", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_DSN_NAME", "unknown")
		t.Setenv("PHOTOPRISM_TEST_DSN_SQLITE", ":memory:?cache=shared&_foreign_keys=on")

		driver, dsn := PhotoPrismTestToDriverDSN()

		assert.Equal(t, "sqlite", driver)
		assert.Equal(t, ":memory:?cache=shared&_foreign_keys=on", dsn)
	})
}
