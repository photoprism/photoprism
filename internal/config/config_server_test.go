package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config/ttl"
)

func TestConfig_HttpServerHost(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "0.0.0.0", c.HttpHost())
	c.options.HttpHost = "test"
	assert.Equal(t, "test", c.HttpHost())
	c.options.HttpHost = "unix:/tmp/photoprism.sock"
	assert.Equal(t, "unix:/tmp/photoprism.sock", c.HttpHost())
}

func TestConfig_HttpSocket(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.HttpSocket())
	c.options.HttpHost = "unix:/tmp/photoprism.sock"
	assert.Equal(t, "/tmp/photoprism.sock", c.HttpSocket())
}

func TestConfig_HttpServerPort(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, 2342, c.HttpPort())
	c.options.HttpPort = 1234
	assert.Equal(t, 1234, c.HttpPort())
}

func TestConfig_HttpServerMode(t *testing.T) {
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

	assert.Equal(t, "index.gohtml", c.TemplateName())
	c.settings.Templates.Default = "rainbow.gohtml"
	assert.Equal(t, "rainbow.gohtml", c.TemplateName())
	c.settings.Templates.Default = "xxx"
	assert.Equal(t, "index.gohtml", c.TemplateName())

}

func TestConfig_HttpCompression(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.HttpCompression())
}

func TestConfig_HttpCachePublic(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.HttpCachePublic())
	c.Options().CdnUrl = "https://cdn.com/"
	assert.True(t, c.HttpCachePublic())
	c.Options().CdnUrl = ""
	assert.False(t, c.HttpCachePublic())
	c.Options().HttpCachePublic = true
	assert.True(t, c.HttpCachePublic())
	c.Options().HttpCachePublic = false
	assert.False(t, c.HttpCachePublic())
}

func TestConfig_HttpCacheMaxAge(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ttl.Duration(2592000), c.HttpCacheMaxAge())
	c.Options().HttpCacheMaxAge = 23
	assert.Equal(t, ttl.Duration(23), c.HttpCacheMaxAge())
	c.Options().HttpCacheMaxAge = 41536000
	assert.Equal(t, ttl.CacheMaxAge, c.HttpCacheMaxAge())
	c.Options().HttpCacheMaxAge = 0
	assert.Equal(t, ttl.Duration(2592000), c.HttpCacheMaxAge())
}

func TestConfig_HttpVideoMaxAge(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ttl.CacheVideo, c.HttpVideoMaxAge())
	c.Options().HttpVideoMaxAge = 23
	assert.Equal(t, ttl.Duration(23), c.HttpVideoMaxAge())
	c.Options().HttpVideoMaxAge = 41536000
	assert.Equal(t, ttl.CacheMaxAge, c.HttpVideoMaxAge())
	c.Options().HttpVideoMaxAge = 0
	assert.Equal(t, ttl.CacheVideo, c.HttpVideoMaxAge())
}
