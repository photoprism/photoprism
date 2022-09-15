package config

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
}
