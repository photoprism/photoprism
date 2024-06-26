package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/oidc"
)

var onceOidc sync.Once

func initOidc() {
	services.OIDC, _ = oidc.NewClient(
		Config().OIDCUri(),
		Config().OIDCClient(),
		Config().OIDCSecret(),
		Config().OIDCScopes(),
		Config().SiteUrl(),
		Config().Debug(),
	)
}

func OIDC() *oidc.Client {
	onceOidc.Do(initOidc)

	return services.OIDC
}
