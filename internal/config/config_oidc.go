package config

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"
)

const OIDCDefaultScopes = "openid email profile"

// OIDCEnabled checks if login via OpenID Connect (OIDC) is enabled.
func (c *Config) OIDCEnabled() bool {
	return c.options.OIDCIssuer != "" && c.options.OIDCClient != "" && c.options.OIDCSecret != ""
}

// OIDCIssuer returns the OpenID Connect Issuer URL as string for single sign-on via OIDC.
func (c *Config) OIDCIssuer() string {
	return c.options.OIDCIssuer
}

// OIDCIssuerURL returns the OpenID Connect Issuer URL as *url.URL for single sign-on via OIDC.
func (c *Config) OIDCIssuerURL() *url.URL {
	if oidcIssuer := c.OIDCIssuer(); oidcIssuer == "" {
		return &url.URL{}
	} else if result, err := url.Parse(oidcIssuer); err != nil {
		log.Errorf("oidc: failed to parse issuer URL (%s)", err)
		return &url.URL{}
	} else {
		return result
	}
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

// OIDCInsecure checks if OIDC issuer SSL/TLS certificate verification should be skipped.
func (c *Config) OIDCInsecure() bool {
	return c.options.OIDCInsecure
}

// OIDCRegister checks if new accounts may be created via OIDC.
func (c *Config) OIDCRegister() bool {
	return c.options.OIDCRegister
}

// OIDCReport returns the OpenID Connect config values as a table for reporting.
func (c *Config) OIDCReport() (rows [][]string, cols []string) {
	cols = []string{"Name", "Value"}

	rows = [][]string{
		{"oidc-issuer", c.OIDCIssuer()},
		{"oidc-client", c.OIDCClient()},
		{"oidc-secret", strings.Repeat("*", utf8.RuneCountInString(c.OIDCSecret()))},
		{"oidc-scopes", c.OIDCScopes()},
		{"oidc-insecure", fmt.Sprintf("%t", c.OIDCInsecure())},
		{"oidc-register", fmt.Sprintf("%t", c.OIDCRegister())},
	}

	return rows, cols
}
