package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// OIDCRedirect creates a new access token for authenticated users and then redirects the browser back to the app.
//
// GET /api/v1/oidc/redirect
func OIDCRedirect(router *gin.RouterGroup) {
	router.GET("/oidc/redirect", func(c *gin.Context) {
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
		action := "redirect"

		// Get global config.
		conf := get.Config()

		// Abort in public mode and if OIDC is disabled.
		if get.Config().Public() {
			event.AuditErr([]string{clientIp, "oidc", actor, action, authn.ErrDisabledInPublicMode.Error()})
			Abort(c, http.StatusForbidden, i18n.ErrForbidden)
			return
		} else if !conf.OIDCEnabled() {
			event.AuditErr([]string{clientIp, "oidc", actor, action, authn.ErrAuthenticationDisabled.Error()})
			Abort(c, http.StatusMethodNotAllowed, i18n.ErrUnsupported)
			return
		}

		// Get OIDC provider.
		provider := get.OIDC()

		if provider == nil {
			event.AuditErr([]string{clientIp, "oidc", actor, action, authn.ErrAuthenticationDisabled.Error()})
			Abort(c, http.StatusInternalServerError, i18n.ErrConnectionFailed)
			return
		}

		_, claimErr := provider.CodeExchangeUserInfo(c)

		if claimErr != nil {
			event.AuditErr([]string{clientIp, "oidc", actor, action, claimErr.Error()})
			Abort(c, http.StatusForbidden, i18n.ErrForbidden)
			return
		}

		// TODO 1: Create user account if it does not exist yet.
		/*
			user := &entity.User{
				DisplayName:  userInfo.GetName(),
				UserName:     oidc.UsernameFromUserInfo(userInfo),
				UserEmail:    userInfo.GetEmail(),
				AuthID:       userInfo.GetSubject(),
				AuthProvider: authn.ProviderOIDC.String(),
			} */

		// TODO 2: Create and return user session.

		// TODO 3: Render HTML template to set the access token in localStorage.

		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": StatusFailed})
	})
}
