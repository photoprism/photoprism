package config

import (
	"testing"

	"github.com/photoprism/photoprism/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	ctx := CliTestContext()

	assert.True(t, ctx.IsSet("assets-path"))
	assert.False(t, ctx.Bool("debug"))

	c := NewConfig(ctx)

	assert.IsType(t, new(Config), c)

	assert.Equal(t, util.ExpandedFilename("../../assets"), c.AssetsPath())
	assert.False(t, c.Debug())
	assert.False(t, c.ReadOnly())
}

func TestConfig_Name(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	name := c.Name()
	assert.Equal(t, "config.test", name)
}

func TestConfig_Version(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	version := c.Version()
	assert.Equal(t, "0.0.0", version)
}

func TestConfig_TensorFlowVersion(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	version := c.TensorFlowVersion()
	assert.Equal(t, "1.14.0", version)
}

func TestConfig_Copyright(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	copyright := c.Copyright()
	assert.Equal(t, "", copyright)
}

func TestConfig_ConfigFile(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	configFile := c.ConfigFile()
	assert.Equal(t, "", configFile)
}

func TestConfig_ConfigPath(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	configPath := c.ConfigPath()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/config", configPath)
}

func TestConfig_PIDFilename(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	filename := c.PIDFilename()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/photoprism.pid", filename)
}

func TestConfig_LogFilename(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	filename := c.LogFilename()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/photoprism.log", filename)
}

func TestConfig_DetachServer(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	detachServer := c.DetachServer()
	assert.Equal(t, false, detachServer)
}

func TestConfig_SqlServerHost(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	host := c.SqlServerHost()
	assert.Equal(t, "127.0.0.1", host)
}

func TestConfig_SqlServerPort(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	port := c.SqlServerPort()
	assert.Equal(t, uint(4000), port)
}

func TestConfig_SqlServerPath(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	path := c.SqlServerPath()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/resources/database", path)
}

func TestConfig_SqlServerPassword(t *testing.T) {
	ctx := CliTestContext()
	c := NewConfig(ctx)

	password := c.SqlServerPassword()
	assert.Equal(t, "", password)
}
