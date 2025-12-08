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

// OAuthAuthorize (placeholder) for Authorization Code Grant consent.
//
//	@Summary	OAuth2 authorization endpoint (not implemented)
//	@Id			OAuthAuthorize
//	@Tags		Authentication
//	@Produce	json
//	@Failure	405	{object}	i18n.Response
//	@Router		/api/v1/oauth/authorize [get]
func OAuthAuthorize(router *gin.RouterGroup) {
	router.GET("/oauth/authorize", func(c *gin.Context) {
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

		// TODO: see https://github.com/photoprism/photoprism/issues/4368

		// Send response.
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": StatusFailed})
	})
}
