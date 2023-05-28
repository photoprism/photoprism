package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_DatabaseDriver(t *testing.T) {
	c := NewConfig(CliTestContext())

	driver := c.DatabaseDriver()
	assert.Equal(t, SQLite3, driver)
}

func TestConfig_ParseDatabaseDsn(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.DatabaseDsn = "foo:b@r@tcp(honeypot:1234)/baz?charset=utf8mb4,utf8&parseTime=true"
	c.options.DatabaseDriver = SQLite3

	assert.Equal(t, "", c.DatabaseServer())
	assert.Equal(t, "", c.DatabaseHost())
	assert.Equal(t, 0, c.DatabasePort())
	assert.Equal(t, "foo:b@r@tcp(honeypot:1234)/baz?charset=utf8mb4,utf8&parseTime=true", c.DatabaseName())
	assert.Equal(t, "", c.DatabaseUser())
	assert.Equal(t, "", c.DatabasePassword())

	c.options.DatabaseDriver = MySQL

	assert.Equal(t, "honeypot:1234", c.DatabaseServer())
	assert.Equal(t, "honeypot", c.DatabaseHost())
	assert.Equal(t, 1234, c.DatabasePort())
	assert.Equal(t, "baz", c.DatabaseName())
	assert.Equal(t, "foo", c.DatabaseUser())
	assert.Equal(t, "b@r", c.DatabasePassword())

	c.options.DatabaseDriver = SQLite3

	assert.Equal(t, "", c.DatabaseServer())
	assert.Equal(t, "", c.DatabaseHost())
	assert.Equal(t, 0, c.DatabasePort())
	assert.Equal(t, "foo:b@r@tcp(honeypot:1234)/baz?charset=utf8mb4,utf8&parseTime=true", c.DatabaseName())
	assert.Equal(t, "", c.DatabaseUser())
	assert.Equal(t, "", c.DatabasePassword())
}

func TestConfig_DatabaseServer(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.DatabaseServer())
	c.options.DatabaseServer = "test"
	assert.Equal(t, "", c.DatabaseServer())
}

func TestConfig_DatabaseHost(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.DatabaseHost())
}

func TestConfig_DatabasePort(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, 0, c.DatabasePort())
}

func TestConfig_DatabasePortString(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.DatabasePortString())
}

func TestConfig_DatabaseName(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/index.db?_busy_timeout=5000", c.DatabaseName())
}

func TestConfig_DatabaseUser(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.DatabaseUser())
}

func TestConfig_DatabasePassword(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.DatabasePassword())
}

func TestConfig_DatabaseDsn(t *testing.T) {
	c := NewConfig(CliTestContext())

	driver := c.DatabaseDriver()
	assert.Equal(t, SQLite3, driver)
	c.options.DatabaseDsn = ""
	c.options.DatabaseDriver = "MariaDB"
	assert.Equal(t, "photoprism:@tcp(localhost)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true", c.DatabaseDsn())
	c.options.DatabaseDriver = "tidb"
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/index.db?_busy_timeout=5000", c.DatabaseDsn())
	c.options.DatabaseDriver = "Postgres"
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/index.db?_busy_timeout=5000", c.DatabaseDsn())
	c.options.DatabaseDriver = "SQLite"
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/index.db?_busy_timeout=5000", c.DatabaseDsn())
	c.options.DatabaseDriver = ""
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/index.db?_busy_timeout=5000", c.DatabaseDsn())
}

func TestConfig_DatabaseFile(t *testing.T) {
	c := NewConfig(CliTestContext())

	driver := c.DatabaseDriver()
	assert.Equal(t, SQLite3, driver)
	c.options.DatabaseDsn = ""
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/index.db", c.DatabaseFile())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/index.db?_busy_timeout=5000", c.DatabaseDsn())
}

func TestConfig_DatabaseConns(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.options.DatabaseConns = 28
	assert.Equal(t, 28, c.DatabaseConns())

	c.options.DatabaseConns = 3000
	assert.Equal(t, 1024, c.DatabaseConns())
}

func TestConfig_DatabaseConnsIdle(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.options.DatabaseConnsIdle = 14
	c.options.DatabaseConns = 28
	assert.Equal(t, 14, c.DatabaseConnsIdle())

	c.options.DatabaseConnsIdle = -55
	assert.Greater(t, c.DatabaseConnsIdle(), 8)

	c.options.DatabaseConnsIdle = 35
	assert.Equal(t, 28, c.DatabaseConnsIdle())
}
