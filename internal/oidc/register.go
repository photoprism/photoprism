package oidc

import (
	"github.com/photoprism/photoprism/internal/config"
)

// init initializes the package.
func init() {
	// Register OpenID Connect extension.
	config.Register("oidc", UpdateConfig, ClientConfig)
}

// ClientConfig returns the OIDC client config values.
func ClientConfig(c *config.Config, t config.ClientType) config.Map {
	result := config.Map{
		"enabled":      c.OIDCEnabled(),
		"redirect":     c.OIDCRedirect(),
		"provider":     c.OIDCProvider(),
		"providerIcon": c.OIDCProviderIcon(),
		"loginUri":     "/api/v1/oidc/login",
	}

	return result
}

// UpdateConfig initializes the OIDC config options.
func UpdateConfig(c *config.Config) error {
	return nil
}
