package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/auth/oidc"
)

var (
	oidcMutex     sync.Mutex
	newOIDCClient = oidc.NewClient
)

// initOIDC initializes and caches the OIDC client if discovery succeeds.
func initOIDC() *oidc.Client {
	client, err := newOIDCClient(
		Config().OIDCUri(),
		Config().OIDCClient(),
		Config().OIDCSecret(),
		Config().OIDCScopes(),
		Config().SiteUrl(),
		false,
	)

	if err != nil {
		return nil
	}

	services.OIDC = client

	return client
}

// resetOIDC clears the cached OIDC client so the next lookup retries initialization.
func resetOIDC() {
	oidcMutex.Lock()
	defer oidcMutex.Unlock()

	services.OIDC = nil
}

// OIDC returns the singleton OIDC client instance.
func OIDC() *oidc.Client {
	oidcMutex.Lock()
	defer oidcMutex.Unlock()

	if services.OIDC != nil {
		return services.OIDC
	}

	return initOIDC()
}
