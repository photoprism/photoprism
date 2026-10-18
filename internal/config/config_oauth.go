package config

import (
	"strings"
)

// OAuthAppsEnabled reports whether apps and MCP clients can be authorized with the OAuth authorization code flow.
func (c *Config) OAuthAppsEnabled() bool {
	if c == nil || c.Public() {
		return false
	}

	switch c.options.OAuthApps {
	case "", "none":
		return false
	default:
		return true
	}
}

// OAuthResourceUrl returns the canonical URL of the MCP endpoint, which is the audience of app access tokens.
func (c *Config) OAuthResourceUrl() string {
	return strings.TrimSuffix(c.SiteUrl(), "/") + ApiUri + "/mcp"
}

// OAuthResourceMetadataUri returns the path of the protected resource metadata for the MCP endpoint.
func (c *Config) OAuthResourceMetadataUri() string {
	return c.BaseUri("/.well-known/oauth-protected-resource" + ApiUri + "/mcp")
}

// OAuthResourceMetadataUrl returns the URL of the protected resource metadata for the MCP endpoint.
func (c *Config) OAuthResourceMetadataUrl() string {
	return c.SiteUrl() + ".well-known/oauth-protected-resource" + ApiUri + "/mcp"
}
