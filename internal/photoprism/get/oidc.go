package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/auth/oidc"
)

var onceOidc sync.Once

func initOidc() {
	services.OIDC, _ = oidc.NewClient(
		Config().OIDCUri(),
		Config().OIDCClient(),
		Config().OIDCSecret(),
		Config().OIDCScopes(),
		Config().SiteUrl(),
		false,
	)
}

func OIDC() *oidc.Client {
	onceOidc.Do(initOidc)

	return services.OIDC
}
