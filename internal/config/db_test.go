package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_DatabaseDriver(t *testing.T) {
	c := NewConfig(CliTestContext())

	driver := c.DatabaseDriver()
	assert.Equal(t, SQLite, driver)
}

func TestConfig_ParseDatabaseDsn(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.params.DatabaseDsn ="foo:b@r@tcp(honeypot:1234)/baz?charset=utf8mb4,utf8&parseTime=true"

	assert.Equal(t, "honeypot:1234", c.DatabaseServer())
	assert.Equal(t, "honeypot", c.DatabaseHost())
	assert.Equal(t, 1234, c.DatabasePort())
	assert.Equal(t, "baz", c.DatabaseName())
	assert.Equal(t, "foo", c.DatabaseUser())
	assert.Equal(t, "b@r", c.DatabasePassword())
}

func TestConfig_DatabaseServer(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "localhost", c.DatabaseServer())
}

func TestConfig_DatabaseHost(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "localhost", c.DatabaseHost())
}

func TestConfig_DatabasePort(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, 3306, c.DatabasePort())
}

func TestConfig_DatabaseName(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "photoprism", c.DatabaseName())
}

func TestConfig_DatabaseUser(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "photoprism", c.DatabaseUser())
}

func TestConfig_DatabasePassword(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.DatabasePassword())
}

func TestConfig_DatabaseDsn(t *testing.T) {
	c := NewConfig(CliTestContext())

	dsn := c.DatabaseDriver()
	assert.Equal(t, SQLite, dsn)
}