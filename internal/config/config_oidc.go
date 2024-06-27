package config

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	OIDCDefaultScopes       = "openid email profile"
	OIDCDefaultProviderName = "OpenID Connect"
	OIDCDefaultProviderIcon = "brands/openid.svg"
	OIDCLoginUri            = ApiUri + "/oidc/login"
	OIDCRedirectUri         = ApiUri + "/oidc/redirect"
)

// OIDCEnabled checks if sign-on via OpenID Connect (OIDC) is fully configured and enabled.
func (c *Config) OIDCEnabled() bool {
	if c.options.DisableOIDC {
		return false
	} else if !c.SiteHttps() {
		// Site URL must start with "https://".
		return false
	} else if !strings.HasPrefix(c.options.OIDCUri, "https://") {
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

// OIDCInsecure checks if OIDC issuer SSL/TLS certificate verification should be skipped.
func (c *Config) OIDCInsecure() bool {
	return c.options.OIDCInsecure
}

// OIDCClient returns the Client ID for single sign-on via OIDC.
func (c *Config) OIDCClient() string {
	return c.options.OIDCClient
}

// OIDCSecret returns the Client ID for single sign-on via OIDC.
func (c *Config) OIDCSecret() string {
	return c.options.OIDCSecret
}

// OIDCScopes returns the user information scopes for single sign-on via OIDC.
func (c *Config) OIDCScopes() string {
	if c.options.OIDCScopes == "" {
		return OIDCDefaultScopes
	}

	return c.options.OIDCScopes
}

// OIDCProvider returns the OIDC provider name.
func (c *Config) OIDCProvider() string {
	if c.options.OIDCProvider == "" {
		return OIDCDefaultProviderName
	}

	return c.options.OIDCProvider
}

// OIDCIcon returns the OIDC provider icon URI.
func (c *Config) OIDCIcon() string {
	if c.options.OIDCIcon == "" {
		return c.StaticAssetUri(OIDCDefaultProviderIcon)
	}

	return c.options.OIDCIcon
}

// OIDCRegister checks if new accounts may be created via OIDC.
func (c *Config) OIDCRegister() bool {
	return c.options.OIDCRegister
}

// OIDCRedirect checks if unauthenticated users should automatically be redirected to the OIDC login page.
func (c *Config) OIDCRedirect() bool {
	return c.options.OIDCRedirect
}

// DisableOIDC checks if single sign-on via OpenID Connect (OIDC) should be disabled.
func (c *Config) DisableOIDC() bool {
	return c.options.DisableOIDC
}

// OIDCLoginUri returns the OIDC login API endpoint URI.
func (c *Config) OIDCLoginUri() string {
	return c.BaseUri(OIDCLoginUri)
}

// OIDCRedirectUri returns the OIDC redirect API endpoint URI.
func (c *Config) OIDCRedirectUri() string {
	return c.BaseUri(OIDCRedirectUri)
}

// OIDCReport returns the OpenID Connect config values as a table for reporting.
func (c *Config) OIDCReport() (rows [][]string, cols []string) {
	cols = []string{"Name", "Value"}

	rows = [][]string{
		{"oidc-uri", c.OIDCUri().String()},
		{"oidc-insecure", fmt.Sprintf("%t", c.OIDCInsecure())},
		{"oidc-client", c.OIDCClient()},
		{"oidc-secret", strings.Repeat("*", utf8.RuneCountInString(c.OIDCSecret()))},
		{"oidc-scopes", c.OIDCScopes()},
		{"oidc-provider", c.OIDCProvider()},
		{"oidc-icon", c.OIDCIcon()},
		{"oidc-register", fmt.Sprintf("%t", c.OIDCRegister())},
		{"oidc-redirect", fmt.Sprintf("%t", c.OIDCRedirect())},
		{"disable-oidc", fmt.Sprintf("%t", c.DisableOIDC())},
	}

	return rows, cols
}
