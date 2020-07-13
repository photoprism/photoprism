package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_HttpServerHost2(t *testing.T) {
	c := NewTestConfig()
	assert.Equal(t, "0.0.0.0", c.HttpServerHost())
	c.params.HttpServerHost = "test"
	assert.Equal(t, "test", c.HttpServerHost())
}

func TestConfig_HttpServerPort2(t *testing.T) {
	c := NewTestConfig()
	assert.Equal(t, int(2342), c.HttpServerPort())
	c.params.HttpServerPort = int(1234)
	assert.Equal(t, int(1234), c.HttpServerPort())
}

func TestConfig_HttpServerMode2(t *testing.T) {
	c := NewTestConfig()
	assert.Equal(t, "debug", c.HttpServerMode())
	c.params.Debug = false
	assert.Equal(t, "release", c.HttpServerMode())
}

func TestConfig_TemplateName(t *testing.T) {
	c := NewTestConfig()
	assert.Equal(t, "index.tmpl", c.TemplateName())
	c.settings.Templates.Default = "rainbow.tmpl"
	assert.Equal(t, "rainbow.tmpl", c.TemplateName())
}
