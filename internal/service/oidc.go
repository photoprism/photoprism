package service

import (
	"sync"

	"github.com/photoprism/photoprism/internal/oidc"
)

var onceOidc sync.Once

func initOidc() {
	services.Oidc = oidc.NewClient(
		Config().OidcIssuerUrl(),
		Config().OidcClientId(),
		Config().OidcClientSecret(),
		Config().SiteUrl(),
		Config().Debug(),
	)
}

func Oidc() *oidc.Client {
	onceOidc.Do(initOidc)
	return services.Oidc
}
