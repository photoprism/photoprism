package service

import (
	"github.com/photoprism/photoprism/internal/oidc"
)

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
	if services.Oidc == nil {
		initOidc()
	}
	return services.Oidc
}
