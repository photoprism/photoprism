package config

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestConfig_OAuthApps(t *testing.T) {
	cases := map[string]string{
		"":                  OAuthAppsRegistered,
		"registered":        OAuthAppsRegistered,
		" Registered ":      OAuthAppsRegistered,
		"any":               OAuthAppsAny,
		"*":                 OAuthAppsAny,
		"none":              OAuthAppsNone,
		"off":               OAuthAppsNone,
		"https://claude.ai": "https://claude.ai/",
		"https://Claude.ai/, https://chatgpt.com": "https://claude.ai/,https://chatgpt.com/",
		"https://claude.ai,https://claude.ai/":    "https://claude.ai/",
		"http://claude.ai":                        OAuthAppsRegistered,
		"https://claude.ai/path":                  OAuthAppsRegistered,
		"foo":                                     OAuthAppsRegistered,
	}
	for value, expected := range cases {
		t.Run(value, func(t *testing.T) {
			c := NewConfig(CliTestContext())
			c.options.OAuthApps = value
			assert.Equal(t, expected, c.OAuthApps())
		})
	}
	t.Run("NilConfig", func(t *testing.T) {
		var c *Config
		assert.Equal(t, OAuthAppsNone, c.OAuthApps())
		assert.Nil(t, c.OAuthAppsOrigins())
	})
}

func TestConfig_OAuthAppsOrigins(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.OAuthApps = "https://claude.ai, http://insecure.example.com, https://chatgpt.com/, https://x.example.com/path"
		assert.Equal(t, []string{"https://claude.ai/", "https://chatgpt.com/"}, c.OAuthAppsOrigins())
	})
	t.Run("Keyword", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.OAuthApps = "any"
		assert.Empty(t, c.OAuthAppsOrigins())
	})
}

func TestConfig_OAuthAppsEnabled(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Public = false
		c.options.OAuthApps = ""
		assert.True(t, c.OAuthAppsEnabled())
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

func TestConfig_OAuthAppsMetadata(t *testing.T) {
	t.Run("Registered", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Public = false
		c.options.OAuthApps = "registered"
		assert.False(t, c.OAuthAppsMetadata("https://claude.ai/oauth/client.json"))
		assert.False(t, c.OAuthAppsMetadataSupported())
	})
	t.Run("Any", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Public = false
		c.options.OAuthApps = "any"
		assert.True(t, c.OAuthAppsMetadata("https://claude.ai/oauth/client.json"))
		assert.False(t, c.OAuthAppsMetadata("http://claude.ai/oauth/client.json"))
		assert.False(t, c.OAuthAppsMetadata("cs5gfen1bgxz7s9i"))
		assert.True(t, c.OAuthAppsMetadataSupported())
	})
	t.Run("Origins", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Public = false
		c.options.OAuthApps = "https://claude.ai"
		assert.True(t, c.OAuthAppsMetadata("https://claude.ai/oauth/client.json"))
		assert.True(t, c.OAuthAppsMetadata("https://CLAUDE.ai/oauth/client.json"))
		assert.False(t, c.OAuthAppsMetadata("https://claude.ai.evil.com/oauth/client.json"))
		assert.False(t, c.OAuthAppsMetadata("https://chatgpt.com/oauth/client.json"))
		assert.True(t, c.OAuthAppsMetadataSupported())
	})
	t.Run("PublicMode", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Public = true
		c.options.OAuthApps = "any"
		assert.False(t, c.OAuthAppsMetadata("https://claude.ai/oauth/client.json"))
		assert.False(t, c.OAuthAppsMetadataSupported())
	})
	t.Run("InvalidClientID", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Public = false
		c.options.OAuthApps = "any"
		assert.False(t, c.OAuthAppsMetadata("https:///client.json"))
		assert.False(t, c.OAuthAppsMetadata("https://u:p@claude.ai/client.json"))
		assert.False(t, c.OAuthAppsMetadata("https://claude.ai"))
		assert.False(t, c.OAuthAppsMetadata("https://claude.ai/"))
		assert.False(t, c.OAuthAppsMetadata("https://claude.ai/client.json?x=1"))
		assert.False(t, c.OAuthAppsMetadata("https://claude.ai/client.json#x"))
	})
	t.Run("NonASCIIHost", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Public = false
		c.options.OAuthApps = "https://claude.ai, https://mcp.idx.example.com"
		assert.True(t, c.OAuthAppsMetadata("https://claude.ai/c.json"))
		assert.False(t, c.OAuthAppsMetadata("https://claude.a\u0130/c.json"))
		assert.False(t, c.OAuthAppsMetadata("https://mcp.\u0130dx.example.com/c.json"))
		assert.False(t, c.OAuthAppsMetadata("https://claude.a\u212a/c.json"))
		assert.False(t, c.OAuthAppsMetadata("https://claude.a%C4%B0/c.json"))
		assert.False(t, c.OAuthAppsMetadata("https://claude.a%c4%b0/c.json"))
		assert.False(t, c.OAuthAppsMetadata("https://claud%65.ai/c.json"))
	})
	t.Run("PathAndQuery", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Public = false
		c.options.OAuthApps = "any"
		assert.False(t, c.OAuthAppsMetadata("https://claude.ai/a/../b.json"))
		assert.False(t, c.OAuthAppsMetadata("https://claude.ai/./b.json"))
		assert.False(t, c.OAuthAppsMetadata("https://claude.ai/b.json?"))
		assert.True(t, c.OAuthAppsMetadata("https://claude.ai/a..b.json"))
	})
}

func TestHasDotSegment(t *testing.T) {
	t.Run("DotSegments", func(t *testing.T) {
		assert.True(t, hasDotSegment("/a/../b"))
		assert.True(t, hasDotSegment("/a/."))
	})
	t.Run("None", func(t *testing.T) {
		assert.False(t, hasDotSegment("/a..b/c.json"))
		assert.False(t, hasDotSegment(""))
	})
}

func TestConfig_OAuthAppsMetadataSupported(t *testing.T) {
	cases := map[string]bool{"registered": false, "none": false, "any": true, "https://claude.ai": true}
	for value, expected := range cases {
		t.Run(value, func(t *testing.T) {
			c := NewConfig(CliTestContext())
			c.options.Public = false
			c.options.OAuthApps = value
			assert.Equal(t, expected, c.OAuthAppsMetadataSupported())
		})
	}
}

func TestConfig_OAuthAppsFlag(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		var found bool

		for _, f := range Flags {
			if f.Name() == "oauth-apps" {
				found = true
				assert.Equal(t, OAuthAppsRegistered, f.Flag.(*cli.StringFlag).Value)
			}
		}

		assert.True(t, found)
	})
	t.Run("Apply", func(t *testing.T) {
		flags := flag.NewFlagSet("test", flag.ContinueOnError)
		flags.String("oauth-apps", "", "doc")
		ctx := cli.NewContext(cli.NewApp(), flags, nil)
		require.NoError(t, ctx.Set("oauth-apps", "none"))

		opt := &Options{}
		require.NoError(t, ApplyCliContext(opt, ctx))
		assert.Equal(t, "none", opt.OAuthApps)
	})
	t.Run("Report", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		rows, _ := c.Report()
		found := false
		for _, row := range rows {
			if row[0] == "oauth-apps" {
				found = true
				assert.Equal(t, OAuthAppsRegistered, row[1])
			}
		}
		assert.True(t, found)
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
