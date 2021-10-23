package service

import (
	"github.com/photoprism/photoprism/internal/oidc"
)

func initOidc() (err error) {
	services.Oidc, err = oidc.NewClient(
		Config().OidcIssuerUrl(),
		Config().OidcClientId(),
		Config().OidcClientSecret(),
		Config().SiteUrl(),
		Config().Debug(),
	)
	return
}

func Oidc() (c *oidc.Client, err error) {
	if services.Oidc == nil {
		err = initOidc()
	}
	return services.Oidc, err
}
