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
		"enabled":  c.OIDCEnabled(),
		"provider": c.OIDCProvider(),
		"icon":     c.OIDCIcon(),
		"register": c.OIDCRegister(),
		"redirect": c.OIDCRedirect(),
		"loginUri": c.OIDCLoginUri(),
	}

	return result
}

// UpdateConfig initializes the OIDC config options.
func UpdateConfig(c *config.Config) error {
	return nil
}
