package config

import (
	"net/url"
	"slices"
	"strings"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/txt"
)

// OAuth app policies, see OAuthApps.
const (
	OAuthAppsRegistered = "registered"
	OAuthAppsAny        = "any"
	OAuthAppsNone       = "none"
)

// OAuthApps returns the policy for connecting apps with OAuth: "registered" (default), "any", "none",
// or a comma-separated list of HTTPS origins from which client metadata documents are accepted.
func (c *Config) OAuthApps() string {
	if c == nil {
		return OAuthAppsNone
	}

	switch s := strings.ToLower(strings.TrimSpace(c.options.OAuthApps)); s {
	case "", OAuthAppsRegistered, "default":
		return OAuthAppsRegistered
	case OAuthAppsAny, "all", "*":
		return OAuthAppsAny
	case OAuthAppsNone, "off", "false", "disabled":
		return OAuthAppsNone
	default:
		if origins := c.OAuthAppsOrigins(); len(origins) > 0 {
			return strings.Join(origins, ",")
		}

		if _, warned := c.warnedOnce.LoadOrStore("oauth-apps: "+s, true); !warned {
			event.SystemWarn([]string{"config", "oauth apps", "unsupported value %s, using %s"}, clean.LogQuote(s), OAuthAppsRegistered)
		}

		return OAuthAppsRegistered
	}
}

// OAuthAppsOrigins returns the HTTPS origins from which client metadata documents are accepted, if any.
func (c *Config) OAuthAppsOrigins() []string {
	if c == nil {
		return nil
	}

	var result []string

	for _, s := range strings.Split(c.options.OAuthApps, ",") {
		s = strings.ToLower(strings.TrimSpace(s))

		if !strings.HasPrefix(s, "https://") || !txt.IsASCII(s) {
			continue
		} else if origin := scheme.OriginURL(s); origin != "" && origin == strings.TrimSuffix(s, "/")+"/" && !slices.Contains(result, origin) {
			result = append(result, origin)
		}
	}

	return result
}

// OAuthAppsEnabled reports whether apps and MCP clients can be authorized with the OAuth authorization code flow.
func (c *Config) OAuthAppsEnabled() bool {
	if c == nil || c.Public() {
		return false
	}

	return c.OAuthApps() != OAuthAppsNone
}

// OAuthAppsMetadata reports whether a client metadata document may be fetched from the specified client ID,
// which must be an ASCII HTTPS URL with a host and a path without dot segments, and without user info, query or fragment.
func (c *Config) OAuthAppsMetadata(clientID string) bool {
	if !c.OAuthAppsEnabled() || !strings.HasPrefix(clientID, "https://") || !txt.IsASCII(clientID) {
		return false
	} else if u, err := url.Parse(clientID); err != nil || u.User != nil || u.Hostname() == "" || strings.Contains(u.Host, "%") ||
		!txt.IsASCII(u.Hostname()) || u.Path == "" || u.Path == "/" || u.RawQuery != "" || u.ForceQuery ||
		strings.Contains(clientID, "#") || hasDotSegment(u.Path) {
		return false
	}

	switch c.OAuthApps() {
	case OAuthAppsAny:
		return true
	case OAuthAppsRegistered:
		return false
	default:
		return slices.Contains(c.OAuthAppsOrigins(), strings.ToLower(scheme.OriginURL(clientID)))
	}
}

// hasDotSegment reports whether a URL path contains a "." or ".." segment.
func hasDotSegment(p string) bool {
	for _, segment := range strings.Split(p, "/") {
		if segment == "." || segment == ".." {
			return true
		}
	}

	return false
}

// OAuthAppsMetadataSupported reports whether client metadata documents are accepted at all, so discovery can advertise them.
func (c *Config) OAuthAppsMetadataSupported() bool {
	if !c.OAuthAppsEnabled() {
		return false
	}

	switch c.OAuthApps() {
	case OAuthAppsRegistered, OAuthAppsNone:
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
