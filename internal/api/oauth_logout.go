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

// OAuthLogout registers the OAuth2/OIDC end-session endpoint for RP-initiated logout.
// The behavior is provided by OAuthLogoutHandler, which defaults to a placeholder and is
// replaced by Portal builds with the OIDC OP end-session handler (see oauth_handlers.go
// and portal/internal/server/register.go). Both GET and POST are registered per the
// OpenID Connect RP-Initiated Logout specification.
//
//	@Summary	OAuth2/OIDC end-session endpoint (RP-initiated logout)
//	@Id			OAuthLogout
//	@Tags		Authentication
//	@Produce	json
//	@Success	302		{string}	string	"redirect to post_logout_redirect_uri"
//	@Failure	403,405	{object}	i18n.Response
//	@Router		/api/v1/oauth/logout [get]
//	@Router		/api/v1/oauth/logout [post]
func OAuthLogout(router *gin.RouterGroup) {
	router.GET("/oauth/logout", func(c *gin.Context) {
		OAuthLogoutHandler(c)
	})
	router.POST("/oauth/logout", func(c *gin.Context) {
		OAuthLogoutHandler(c)
	})
}

// defaultOAuthLogout is the placeholder end-session handler used unless a Portal build
// overrides OAuthLogoutHandler with the OIDC OP end-session handler.
func defaultOAuthLogout(c *gin.Context) {
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
	action := "logout"

	// Abort if running in public mode.
	if get.Config().Public() {
		event.AuditErr([]string{clientIp, "oauth2", actor, action, authn.ErrDisabledInPublicMode.Error()})
		Abort(c, http.StatusForbidden, i18n.ErrForbidden)
		return
	}

	// The OIDC OP end-session endpoint is only available on Portal builds.
	c.JSON(http.StatusMethodNotAllowed, gin.H{"status": StatusFailed})
}
