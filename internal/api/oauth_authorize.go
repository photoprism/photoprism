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

// OAuthAuthorize registers the OAuth2/OIDC authorization endpoint. The behavior
// is provided by OAuthAuthorizeHandler, which defaults to a placeholder and is
// replaced by Portal builds with the OIDC OP authorize flow (see
// oauth_handlers.go and portal/internal/server/register.go).
//
//	@Summary	OAuth2/OIDC authorization endpoint
//	@Id			OAuthAuthorize
//	@Tags		Authentication
//	@Produce	json
//	@Success	302		{string}	string	"redirect to the client redirect_uri with an authorization code"
//	@Failure	403,405	{object}	i18n.Response
//	@Router		/api/v1/oauth/authorize [get]
func OAuthAuthorize(router *gin.RouterGroup) {
	router.GET("/oauth/authorize", func(c *gin.Context) {
		OAuthAuthorizeHandler(c)
	})
}

// defaultOAuthAuthorize is the placeholder authorization handler used unless a
// Portal build overrides OAuthAuthorizeHandler with the OIDC OP authorize flow.
func defaultOAuthAuthorize(c *gin.Context) {
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
	action := "authorize"

	// Abort if running in public mode.
	if get.Config().Public() {
		event.AuditErr([]string{clientIp, "oauth2", actor, action, authn.ErrDisabledInPublicMode.Error()})
		Abort(c, http.StatusForbidden, i18n.ErrForbidden)
		return
	}

	// The OIDC OP authorization endpoint is only available on Portal builds.
	c.JSON(http.StatusMethodNotAllowed, gin.H{"status": StatusFailed})
}
