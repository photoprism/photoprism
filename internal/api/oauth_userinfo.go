package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// OAuthUserinfo registers the OAuth2/OIDC userinfo endpoint. The behavior is
// provided by OAuthUserinfoHandler, which defaults to a placeholder and is
// replaced by Portal builds with the OIDC OP userinfo handler (see
// oauth_handlers.go and portal/internal/server/register.go). Both GET and POST
// are registered per OIDC Core §5.3.1.
//
//	@Summary	OAuth2/OIDC userinfo endpoint
//	@Id			OAuthUserinfo
//	@Tags		Authentication
//	@Produce	json
//	@Success	200		{object}	gin.H
//	@Failure	401,405	{object}	i18n.Response
//	@Router		/api/v1/oauth/userinfo [get]
//	@Router		/api/v1/oauth/userinfo [post]
func OAuthUserinfo(router *gin.RouterGroup) {
	router.GET("/oauth/userinfo", func(c *gin.Context) {
		OAuthUserinfoHandler(c)
	})
	router.POST("/oauth/userinfo", func(c *gin.Context) {
		OAuthUserinfoHandler(c)
	})
}

// defaultOAuthUserinfo is the placeholder userinfo handler used unless a Portal
// build overrides OAuthUserinfoHandler with the OIDC OP userinfo handler.
func defaultOAuthUserinfo(c *gin.Context) {
	// Prevent CDNs from caching this endpoint.
	if header.IsCdn(c.Request) {
		AbortNotFound(c)
		return
	}

	// Disable caching of responses.
	c.Header(header.CacheControl, header.CacheControlNoStore)

	// Get client IP address for logs and rate limiting checks.
	clientIp := ClientIP(c)
	actor := "unknown client"
	action := "userinfo"

	// Abort if running in public mode.
	if get.Config().Public() {
		event.AuditErr([]string{clientIp, "oauth2", actor, action, authn.ErrDisabledInPublicMode.Error()})
		Abort(c, http.StatusForbidden, i18n.ErrForbidden)
		return
	}

	// The OIDC OP userinfo endpoint is only available on Portal builds.
	c.JSON(http.StatusMethodNotAllowed, gin.H{"status": StatusFailed})
}
