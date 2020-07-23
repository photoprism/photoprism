package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_HttpServerHost2(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "0.0.0.0", c.HttpServerHost())
	c.params.HttpServerHost = "test"
	assert.Equal(t, "test", c.HttpServerHost())
}

func TestConfig_HttpServerPort2(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, int(2342), c.HttpServerPort())
	c.params.HttpServerPort = int(1234)
	assert.Equal(t, int(1234), c.HttpServerPort())
}

func TestConfig_HttpServerMode2(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "release", c.HttpServerMode())
	c.params.Debug = true
	assert.Equal(t, "debug", c.HttpServerMode())
	c.params.HttpServerMode = "test"
	assert.Equal(t, "test", c.HttpServerMode())
}

func TestConfig_TemplateName(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "index.tmpl", c.TemplateName())
	c.settings.Templates.Default = "rainbow.tmpl"
	assert.Equal(t, "rainbow.tmpl", c.TemplateName())
	c.settings.Templates.Default = "xxx"
	assert.Equal(t, "index.tmpl", c.TemplateName())

}
