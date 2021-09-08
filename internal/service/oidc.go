package service

import (
	"github.com/photoprism/photoprism/internal/oidc"
	"sync"
)

var onceOidc sync.Once

func initOidc() {
	services.Oidc = oidc.NewClient(
		Config().OidcIssuerUrl(),
		Config().OidcClientId(),
		Config().OidcClientSecret(),
		Config().SiteUrl(),
	)
}

func Oidc() *oidc.Client {
	onceOidc.Do(initOidc)
	return services.Oidc
}
