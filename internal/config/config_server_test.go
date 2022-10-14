package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_HttpServerHost2(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "0.0.0.0", c.HttpHost())
	c.options.HttpHost = "test"
	assert.Equal(t, "test", c.HttpHost())
}

func TestConfig_HttpServerPort2(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, int(2342), c.HttpPort())
	c.options.HttpPort = int(1234)
	assert.Equal(t, int(1234), c.HttpPort())
}

func TestConfig_HttpServerMode2(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, HttpModeProd, c.HttpMode())
	c.options.Debug = true
	assert.Equal(t, HttpModeDebug, c.HttpMode())
	c.options.HttpMode = "test"
	assert.Equal(t, "test", c.HttpMode())
}

func TestConfig_TemplateName(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.initSettings()

	assert.Equal(t, "index.tmpl", c.TemplateName())
	c.settings.Templates.Default = "rainbow.tmpl"
	assert.Equal(t, "rainbow.tmpl", c.TemplateName())
	c.settings.Templates.Default = "xxx"
	assert.Equal(t, "index.tmpl", c.TemplateName())

}

func TestConfig_HttpCompression(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.HttpCompression())
}
