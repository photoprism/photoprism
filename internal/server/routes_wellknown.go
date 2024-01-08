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
	// Registers the "/.well-known/oauth-authorization-server" service discovery endpoint for OAuth2 clients.
	router.Any(conf.BaseUri("/.well-known/oauth-authorization-server"), func(c *gin.Context) {
		response := &wellknown.OAuthAuthorizationServer{
			Issuer:                                    conf.SiteUrl(),
			AuthorizationEndpoint:                     "",
			TokenEndpoint:                             fmt.Sprintf("%sapi/v1/oauth/token", conf.SiteUrl()),
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
			EndSessionEndpoint:                        fmt.Sprintf("%sapi/v1/oauth/logout", conf.SiteUrl()),
			RequestParameterSupported:                 false,
			RequestObjectSigningAlgValuesSupported:    []string{},
			DeviceAuthorizationEndpoint:               "",
			DpopSigningAlgValuesSupported:             []string{},
		}

		c.JSON(http.StatusOK, response)
	})

	// Registers the "/.well-known/openid-configuration" service discovery endpoint for OpenID Connect clients.
	router.Any(conf.BaseUri("/.well-known/openid-configuration"), func(c *gin.Context) {
		response := &wellknown.OpenIDConfiguration{
			Issuer:                                    conf.SiteUrl(),
			AuthorizationEndpoint:                     "",
			TokenEndpoint:                             fmt.Sprintf("%sapi/v1/oauth/token", conf.SiteUrl()),
			UserinfoEndpoint:                          "",
			RegistrationEndpoint:                      "",
			JwksUri:                                   "",
			ResponseTypesSupported:                    []string{"token"},
			ResponseModesSupported:                    []string{},
			GrantTypesSupported:                       []string{"client_credentials"},
			SubjectTypesSupported:                     []string{},
			IdTokenSigningAlgValuesSupported:          []string{},
			ScopesSupported:                           acl.Resources.Resources(),
			TokenEndpointAuthMethodsSupported:         []string{"client_secret_basic", "client_secret_post"},
			ClaimsSupported:                           []string{},
			CodeChallengeMethodsSupported:             []string{},
			IntrospectionEndpoint:                     "",
			IntrospectionEndpointAuthMethodsSupported: []string{},
			RevocationEndpoint:                        "",
			RevocationEndpointAuthMethodsSupported:    []string{},
			EndSessionEndpoint:                        fmt.Sprintf("%sapi/v1/oauth/logout", conf.SiteUrl()),
			RequestParameterSupported:                 false,
			RequestObjectSigningAlgValuesSupported:    []string{},
			DeviceAuthorizationEndpoint:               "",
			DpopSigningAlgValuesSupported:             []string{},
		}

		c.JSON(http.StatusOK, response)
	})

}
