package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// OIDCLogin redirects a browser to the login page of the configured OpenID Connect provider, if any.
//
// GET /api/v1/oidc/login
func OIDCLogin(router *gin.RouterGroup) {
	router.GET("/oidc/login", func(c *gin.Context) {
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			AbortNotFound(c)
			return
		}

		// Disable caching of responses.
		c.Header(header.CacheControl, header.CacheControlNoStore)

		// Get client IP address for logs and rate limiting checks.
		clientIp := ClientIP(c)
		action := "sign in"

		// Get global config.
		conf := get.Config()

		// Abort in public mode and if OIDC is disabled.
		if get.Config().Public() {
			event.AuditErr([]string{clientIp, "oidc", action, authn.ErrDisabledInPublicMode.Error()})
			c.Redirect(http.StatusTemporaryRedirect, conf.LoginUri())
			return
		} else if !conf.OIDCEnabled() {
			event.AuditErr([]string{clientIp, "oidc", action, authn.ErrAuthenticationDisabled.Error()})
			c.Redirect(http.StatusTemporaryRedirect, conf.LoginUri())
			return
		}

		// Check request rate limit.
		var r *limiter.Request
		r = limiter.Login.Request(clientIp)

		// Abort if failure rate limit is exceeded.
		if r.Reject() || limiter.Auth.Reject(clientIp) {
			c.HTML(http.StatusTooManyRequests, "auth.gohtml", CreateSessionError(http.StatusTooManyRequests, i18n.Error(i18n.ErrTooManyRequests)))
			return
		}

		// Get OIDC provider.
		provider := get.OIDC()

		if provider == nil {
			event.AuditErr([]string{clientIp, "oidc", action, authn.ErrInvalidProviderConfiguration.Error()})
			c.HTML(http.StatusInternalServerError, "auth.gohtml", CreateSessionError(http.StatusInternalServerError, i18n.Error(i18n.ErrConnectionFailed)))
			return
		}

		// Return the reserved request rate limit token.
		r.Success()

		// Handle OIDC login request.
		provider.AuthCodeUrlHandler(c)
	})
}
