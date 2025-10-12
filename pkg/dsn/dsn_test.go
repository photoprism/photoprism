package dsn

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDSN(t *testing.T) {
	t.Run("File", func(t *testing.T) {
		dsn := NewDSN("/go/src/github.com/photoprism/photoprism/storage/index.db?_busy_timeout=5000")

		assert.Equal(t, "", dsn.Driver)
		assert.Equal(t, "", dsn.User)
		assert.Equal(t, "", dsn.Password)
		assert.Equal(t, "", dsn.Net)
		assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage", dsn.Server)
		assert.Equal(t, "index.db", dsn.Name)
		assert.Equal(t, "_busy_timeout=5000", dsn.Params)
	})
	t.Run("Server", func(t *testing.T) {
		dsn := NewDSN(fmt.Sprintf(
			"%s:%s@%s/%s?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
			"root",
			"FooBar23!",
			"127.0.0.1:3306",
			"test",
		))

		assert.Equal(t, "", dsn.Driver)
		assert.Equal(t, "root", dsn.User)
		assert.Equal(t, "FooBar23!", dsn.Password)
		assert.Equal(t, "", dsn.Net)
		assert.Equal(t, "127.0.0.1:3306", dsn.Server)
		assert.Equal(t, "test", dsn.Name)
		assert.Equal(t, "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true", dsn.Params)
	})
	t.Run("Driver", func(t *testing.T) {
		dsn := NewDSN("mysql://john:pass@localhost:3306/my_db")

		assert.Equal(t, "mysql", dsn.Driver)
		assert.Equal(t, "john", dsn.User)
		assert.Equal(t, "pass", dsn.Password)
		assert.Equal(t, "", dsn.Net)
		assert.Equal(t, "localhost:3306", dsn.Server)
		assert.Equal(t, "my_db", dsn.Name)
		assert.Equal(t, "", dsn.Params)
	})
	t.Run("Net", func(t *testing.T) {
		dsn := NewDSN("mysql://john:pass@tcp(localhost:3306)/my_db")

		assert.Equal(t, "mysql", dsn.Driver)
		assert.Equal(t, "john", dsn.User)
		assert.Equal(t, "pass", dsn.Password)
		assert.Equal(t, "tcp", dsn.Net)
		assert.Equal(t, "localhost:3306", dsn.Server)
		assert.Equal(t, "my_db", dsn.Name)
		assert.Equal(t, "", dsn.Params)
	})

	t.Run("PostgreSQL URI 1", func(t *testing.T) {
		dsn := NewDSN("postgresql://john:pass@postgres:5432/my_db?TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable")

		assert.Equal(t, "postgresql", dsn.Driver)
		assert.Equal(t, "john", dsn.User)
		assert.Equal(t, "pass", dsn.Password)
		assert.Equal(t, "", dsn.Net)
		assert.Equal(t, "postgres:5432", dsn.Server)
		assert.Equal(t, "my_db", dsn.Name)
		assert.Equal(t, "TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable", dsn.Params)
	})

	t.Run("PostgreSQL URI 2", func(t *testing.T) {
		dsn := NewDSN("postgres://john:pass@postgres:5432/my_db?TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable")

		assert.Equal(t, "postgres", dsn.Driver)
		assert.Equal(t, "john", dsn.User)
		assert.Equal(t, "pass", dsn.Password)
		assert.Equal(t, "", dsn.Net)
		assert.Equal(t, "postgres:5432", dsn.Server)
		assert.Equal(t, "my_db", dsn.Name)
		assert.Equal(t, "TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable", dsn.Params)
	})

	t.Run("PostgreSQL Keywords", func(t *testing.T) {
		dsn := NewDSN("host=postgres port=5432 dbname=my_db user=john password=pass connect_timeout=15 sslmode=disable TimeZone=UTC application_name='Photo Prism'")

		assert.Equal(t, "postgresql", dsn.Driver)
		assert.Equal(t, "john", dsn.User)
		assert.Equal(t, "pass", dsn.Password)
		assert.Equal(t, "", dsn.Net)
		assert.Equal(t, "postgres:5432", dsn.Server)
		assert.Equal(t, "my_db", dsn.Name)
		assert.Contains(t, dsn.Params, "connect_timeout=15")
		assert.Contains(t, dsn.Params, "sslmode=disable")
		assert.Contains(t, dsn.Params, "TimeZone=UTC")
		assert.Contains(t, dsn.Params, "application_name=%27Photo+Prism%27")
	})
}
