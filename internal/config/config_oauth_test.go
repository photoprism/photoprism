package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_OAuthAppsEnabled(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Public = false
		c.options.OAuthApps = ""
		assert.False(t, c.OAuthAppsEnabled())
	})
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Public = false
		c.options.OAuthApps = "none"
		assert.False(t, c.OAuthAppsEnabled())
	})
	t.Run("Registered", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Public = false
		c.options.OAuthApps = "registered"
		assert.True(t, c.OAuthAppsEnabled())
	})
	t.Run("PublicMode", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Public = true
		c.options.OAuthApps = "registered"
		assert.False(t, c.OAuthAppsEnabled())
	})
	t.Run("NilConfig", func(t *testing.T) {
		var c *Config
		assert.False(t, c.OAuthAppsEnabled())
	})
}

func TestConfig_OAuthResourceUrl(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.SiteUrl = "https://photos.example.com/"
		assert.Equal(t, "https://photos.example.com/api/v1/mcp", c.OAuthResourceUrl())
		assert.Equal(t, "/.well-known/oauth-protected-resource/api/v1/mcp", c.OAuthResourceMetadataUri())
		assert.Equal(t, "https://photos.example.com/.well-known/oauth-protected-resource/api/v1/mcp", c.OAuthResourceMetadataUrl())
	})
	t.Run("BasePath", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.SiteUrl = "https://example.com/photos/"
		assert.Equal(t, "https://example.com/photos/api/v1/mcp", c.OAuthResourceUrl())
		assert.Equal(t, "/photos/.well-known/oauth-protected-resource/api/v1/mcp", c.OAuthResourceMetadataUri())
		assert.Equal(t, "https://example.com/photos/.well-known/oauth-protected-resource/api/v1/mcp", c.OAuthResourceMetadataUrl())
	})
}
