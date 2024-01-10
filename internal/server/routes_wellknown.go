package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/server/wellknown"
)

// registerWellknownRoutes configures the ".well-known" service discovery routes.
func registerWellknownRoutes(router *gin.Engine, conf *config.Config) {
	// Registers the "/.well-known/oauth-authorization-server" service discovery endpoint for OAuth2 clients.
	router.Any(conf.BaseUri("/.well-known/oauth-authorization-server"), func(c *gin.Context) {
		c.JSON(http.StatusOK, wellknown.NewOAuthAuthorizationServer(conf))
	})

	// Registers the "/.well-known/openid-configuration" service discovery endpoint for OpenID Connect clients.
	router.Any(conf.BaseUri("/.well-known/openid-configuration"), func(c *gin.Context) {
		c.JSON(http.StatusOK, wellknown.NewOpenIDConfiguration(conf))
	})
}
