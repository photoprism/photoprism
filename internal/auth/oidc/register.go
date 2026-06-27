package oidc

import (
	"github.com/photoprism/photoprism/internal/config"
)

// init initializes the package.
func init() {
	// Register OpenID Connect extension.
	config.Register(config.StageInit, "oidc", InitConfig, ClientConfig)
}

// ClientConfig returns the OIDC client config values.
func ClientConfig(c *config.Config, t config.ClientType) config.Values {
	result := config.Values{
		"enabled":  c.OIDCEnabled(),
		"provider": c.OIDCProvider(),
		"icon":     c.OIDCIcon(),
		"register": c.OIDCRegister(),
		"redirect": c.OIDCRedirect(),
		// logout lets the UI know sign-out will redirect to the provider's end-session
		// endpoint; the backend still drives the actual redirect URL.
		"logout":   c.OIDCLogout(),
		"loginUri": c.OIDCLoginUri(),
		// cluster lets the frontend route a cluster-OIDC sign-out to the Portal login.
		"cluster": c.ClusterOIDC(),
		// portalLoginUri is the Portal's browser-facing login page (persisted from
		// the register response), so sign-out lands there without re-initiating
		// the instance OIDC roundtrip and pinning a return_to.
		"portalLoginUri": c.PortalLoginUrl(),
	}

	return result
}

// InitConfig initializes the OIDC config options.
func InitConfig(c *config.Config) error {
	return nil
}
