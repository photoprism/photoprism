package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/oidc"
)

var onceOidc sync.Once

func initOidc() {
	services.OIDC, _ = oidc.NewClient(
		Config().OIDCIssuerURL(),
		Config().OIDCClient(),
		Config().OIDCSecret(),
		Config().OIDCScopes(),
		Config().SiteUrl(),
		Config().Debug(),
	)
}

func OIDC() *oidc.Client {
	oncePhotos.Do(initOidc)

	return services.OIDC
}
