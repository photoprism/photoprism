package config

const OIDCDefaultScopes = "openid profile"

// OIDCEnabled checks if login via OpenID Connect (OIDC) is enabled.
func (c *Config) OIDCEnabled() bool {
	return c.options.OIDCIssuer != "" && c.options.OIDCClient != "" && c.options.OIDCSecret != ""
}

// OIDCIssuer returns the OpenID Connect Issuer URL for single sign-on via OIDC.
func (c *Config) OIDCIssuer() string {
	return c.options.OIDCIssuer
}

// OIDCClient returns the Client ID for single sign-on via OIDC.
func (c *Config) OIDCClient() string {
	return c.options.OIDCClient
}

// OIDCSecret returns the Client ID for single sign-on via OIDC.
func (c *Config) OIDCSecret() string {
	return c.options.OIDCSecret
}

// OIDCScopes returns the token request scopes for single sign-on via OIDC.
func (c *Config) OIDCScopes() string {
	return c.options.OIDCScopes
}

// OIDCInsecure checks if OIDC issuer SSL/TLS certificate verification should be skipped.
func (c *Config) OIDCInsecure() bool {
	return c.options.OIDCInsecure
}
