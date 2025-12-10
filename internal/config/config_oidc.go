package config

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const (
	// OidcDefaultProviderName is the default display name for the built-in OIDC provider.
	OidcDefaultProviderName = "OpenID"
	// OidcDefaultProviderIcon is the default icon path for the built-in OIDC provider.
	OidcDefaultProviderIcon = "img/oidc.svg"
	// OidcLoginUri is the login endpoint path for OIDC.
	OidcLoginUri = ApiUri + "/oidc/login"
	// OidcRedirectUri is the callback endpoint path for OIDC.
	OidcRedirectUri = ApiUri + "/oidc/redirect"
)

// OIDCEnabled checks if sign-on via OpenID Connect (OIDC) is fully configured and enabled.
func (c *Config) OIDCEnabled() bool {
	switch {
	case c.options.DisableOIDC:
		return false
	case !c.SiteHttps():
		// Site URL must start with "https://".
		return false
	case !strings.HasPrefix(c.options.OIDCUri, "https://"):
		// OIDC provider URI must start with "https://".
		return false
	}

	return c.options.OIDCClient != "" && c.options.OIDCSecret != ""
}

// OIDCUri returns the OpenID Connect provider URI as *url.URL for single sign-on via OIDC.
func (c *Config) OIDCUri() *url.URL {
	if uri := c.options.OIDCUri; uri == "" {
		return &url.URL{}
	} else if result, err := url.Parse(uri); err != nil {
		log.Warnf("oidc: failed to parse provider URI (%s)", err)
		return &url.URL{}
	} else if result.Scheme == "https" {
		return result
	} else {
		log.Warnf("oidc: insecure or unsupported provider URI (%s)", uri)
		return &url.URL{}
	}
}

// OIDCClient returns the Client ID for single sign-on via OIDC.
func (c *Config) OIDCClient() string {
	return c.options.OIDCClient
}

// OIDCSecret returns the Client Secret for single sign-on via OIDC.
func (c *Config) OIDCSecret() string {
	// Try to read secret from file if c.options.OIDCSecret is not set.
	if c.options.OIDCSecret != "" {
		return clean.Password(c.options.OIDCSecret)
	} else if fileName := FlagFilePath("OIDC_SECRET"); fileName == "" {
		// No secret set, this is not an error.
		return ""
	} else if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 { //nolint:gosec // path derived from config directory
		log.Warnf("config: failed to read OIDC client secret from %s (%s)", fileName, err)
		return ""
	} else {
		return clean.Password(string(b))
	}
}

// OIDCScopes returns the user information scopes for single sign-on via OIDC.
func (c *Config) OIDCScopes() string {
	if c.options.OIDCScopes == "" {
		return authn.OidcDefaultScopes
	}

	return c.options.OIDCScopes
}

// OIDCProvider returns the OIDC provider name.
func (c *Config) OIDCProvider() string {
	if c.options.OIDCProvider == "" {
		return OidcDefaultProviderName
	}

	return c.options.OIDCProvider
}

// OIDCIcon returns the OIDC provider icon URI.
func (c *Config) OIDCIcon() string {
	if c.options.OIDCIcon != "" {
		if themeIcon := filepath.Join(c.ThemePath(), c.options.OIDCIcon); fs.FileExistsNotEmpty(themeIcon) {
			return path.Join(ThemeUri, c.options.OIDCIcon)
		}

		return c.options.OIDCIcon
	}

	return c.StaticAssetUri(OidcDefaultProviderIcon)
}

// OIDCRedirect checks if unauthenticated users should automatically be redirected to the OIDC login page.
func (c *Config) OIDCRedirect() bool {
	return c.options.OIDCRedirect
}

// OIDCRegister checks if new accounts may be created via OIDC.
func (c *Config) OIDCRegister() bool {
	return c.options.OIDCRegister
}

// OIDCUsername returns the preferred username claim for new users signing up via OIDC.
func (c *Config) OIDCUsername() string {
	switch c.options.OIDCUsername {
	case authn.OidcClaimName:
		return authn.OidcClaimName
	case authn.OidcClaimNickname:
		return authn.OidcClaimNickname
	case authn.OidcClaimEmail:
		return authn.OidcClaimEmail
	default:
		return authn.OidcClaimPreferredUsername
	}
}

// OIDCGroupClaim returns the claim name that should contain security group identifiers.
func (c *Config) OIDCGroupClaim() string {
	if claim := strings.TrimSpace(c.options.OIDCGroupClaim); claim != "" {
		return claim
	}

	return "groups"
}

// OIDCGroup returns the normalized list of required groups; empty means no group check.
func (c *Config) OIDCGroup() []string {
	if len(c.options.OIDCGroup) == 0 {
		return nil
	}

	result := make([]string, 0, len(c.options.OIDCGroup))

	for _, g := range c.options.OIDCGroup {
		if n := normalizeGroupID(g); n != "" {
			result = append(result, n)
		}
	}

	return result
}

// OIDCGroupRoles maps normalized group identifiers to roles.
func (c *Config) OIDCGroupRoles() map[string]acl.Role {
	result := make(map[string]acl.Role, len(c.options.OIDCGroupRole))

	for _, entry := range c.options.OIDCGroupRole {
		entry = strings.TrimSpace(entry)

		if entry == "" {
			continue
		}

		sep := strings.IndexAny(entry, "=:")

		if sep < 1 || sep >= len(entry)-1 {
			continue
		}

		group := normalizeGroupID(entry[:sep])
		role := acl.ParseRole(entry[sep+1:])

		if group == "" || role == acl.RoleNone {
			continue
		}

		result[group] = role
	}

	return result
}

// OIDCDomain returns the email domain name for restricted single sign-on via OIDC.
func (c *Config) OIDCDomain() string {
	return clean.Domain(c.options.OIDCDomain)
}

// OIDCRole returns the default user role when signing up via OIDC.
func (c *Config) OIDCRole() acl.Role {
	if c.options.OIDCRole == "" {
		return acl.RoleGuest
	}

	role := acl.UserRoles[clean.Role(c.options.OIDCRole)]

	if role != acl.RoleNone {
		return role
	}

	return acl.RoleNone
}

// OIDCWebDAV checks if newly registered accounts should be allowed to use WebDAV if their role allows.
func (c *Config) OIDCWebDAV() bool {
	return c.options.OIDCWebDAV
}

// DisableOIDC checks if single sign-on via OpenID Connect (OIDC) should be disabled.
func (c *Config) DisableOIDC() bool {
	return c.options.DisableOIDC
}

// OIDCLoginUri returns the OIDC login API endpoint URI.
func (c *Config) OIDCLoginUri() string {
	return c.BaseUri(OidcLoginUri)
}

// OIDCRedirectUri returns the OIDC redirect API endpoint URI.
func (c *Config) OIDCRedirectUri() string {
	return c.BaseUri(OidcRedirectUri)
}

// OIDCReport returns the OpenID Connect config values as a table for reporting.
func (c *Config) OIDCReport() (rows [][]string, cols []string) {
	cols = []string{"Name", "Value"}

	rows = [][]string{
		{"oidc-uri", c.OIDCUri().String()},
		{"oidc-client", c.OIDCClient()},
		{"oidc-secret", strings.Repeat("*", utf8.RuneCountInString(c.OIDCSecret()))},
		{"oidc-scopes", c.OIDCScopes()},
		{"oidc-provider", c.OIDCProvider()},
		{"oidc-icon", c.OIDCIcon()},
		{"oidc-redirect", fmt.Sprintf("%t", c.OIDCRedirect())},
		{"oidc-register", fmt.Sprintf("%t", c.OIDCRegister())},
		{"oidc-username", c.OIDCUsername()},
	}

	if domain := c.OIDCDomain(); domain != "" {
		rows = append(rows, []string{"oidc-domain", domain})
	}

	if claim := c.OIDCGroupClaim(); claim != "" {
		rows = append(rows, []string{"oidc-group-claim", claim})
	}

	if groups := c.OIDCGroup(); len(groups) > 0 {
		rows = append(rows, []string{"oidc-group", strings.Join(groups, ",")})
	}

	if roles := c.OIDCGroupRoles(); len(roles) > 0 {
		pairs := make([]string, 0, len(roles))

		for g, r := range roles {
			pairs = append(pairs, fmt.Sprintf("%s=%s", g, r))
		}

		sort.Strings(pairs)
		rows = append(rows, []string{"oidc-group-role", strings.Join(pairs, ",")})
	}

	rows = append(rows, [][]string{
		{"oidc-role", c.OIDCRole().String()},
		{"oidc-webdav", fmt.Sprintf("%t", c.OIDCWebDAV())},
		{"disable-oidc", fmt.Sprintf("%t", c.DisableOIDC())},
	}...)

	return rows, cols
}

// normalizeGroupID lowercases and sanitizes a group identifier (GUID or name) for comparisons.
func normalizeGroupID(id string) string {
	return strings.ToLower(clean.Auth(id))
}
