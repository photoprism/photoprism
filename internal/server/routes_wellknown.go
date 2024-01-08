package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/server/wellknown"
)

// registerWellknownRoutes configures the ".well-known" service discovery routes.
func registerWellknownRoutes(router *gin.Engine, conf *config.Config) {
	router.Any(conf.BaseUri("/.well-known/oauth-authorization-server"), func(c *gin.Context) {
		response := &wellknown.OAuthAuthorizationServer{
			Issuer:                                    conf.SiteUrl(),
			TokenEndpoint:                             fmt.Sprintf("%sapi/v1/oauth/token", conf.SiteUrl()),
			EndSessionEndpoint:                        fmt.Sprintf("%sapi/v1/oauth/logout", conf.SiteUrl()),
			ScopesSupported:                           acl.Resources.Resources(),
			ResponseTypesSupported:                    []string{"token"},
			GrantTypesSupported:                       []string{"client_credentials"},
			TokenEndpointAuthMethodsSupported:         []string{"client_secret_basic", "client_secret_post"},
			ResponseModesSupported:                    []string{},
			SubjectTypesSupported:                     []string{},
			ClaimsSupported:                           []string{},
			CodeChallengeMethodsSupported:             []string{},
			IntrospectionEndpointAuthMethodsSupported: []string{},
			RevocationEndpointAuthMethodsSupported:    []string{},
			RequestParameterSupported:                 false,
			RequestObjectSigningAlgValuesSupported:    []string{},
			DeviceAuthorizationEndpoint:               "",
			DpopSigningAlgValuesSupported:             []string{},
		}

		c.JSON(http.StatusOK, response)
	})
}
