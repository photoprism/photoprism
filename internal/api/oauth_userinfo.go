package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// OAuthUserinfo returns standard OpenID Connect UserInfo claims for the account
// the request's access token belongs to, see
// https://github.com/photoprism/photoprism/issues/4369.
//
// Authentication uses the regular PhotoPrism auth/session token (header Bearer
// or X-Auth-Token), so existing clients and app passwords work without an
// OpenID-specific access token; claims are populated from the User entity.
//
//	@Summary	returns information about the authenticated user
//	@Id			OAuthUserinfo
//	@Tags		Authentication
//	@Produce	json
//	@Success	200			{object}	oidc.UserInfo
//	@Failure	401,403		{object}	i18n.Response
//	@Router		/api/v1/oauth/userinfo [get]
//	@Router		/api/v1/oauth/userinfo [post]
func OAuthUserinfo(router *gin.RouterGroup) {
	handler := func(c *gin.Context) {
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			AbortNotFound(c)
			return
		}

		// Disable caching of responses.
		c.Header(header.CacheControl, header.CacheControlNoStore)

		// Get client IP address for logs and rate limiting checks.
		clientIp := ClientIP(c)
		action := "userinfo"

		// Abort if running in public mode.
		if get.Config().Public() {
			event.AuditErr([]string{clientIp, "oauth2", "unknown client", action, authn.ErrDisabledInPublicMode.Error()})
			Abort(c, http.StatusForbidden, i18n.ErrForbidden)
			return
		}

		// Resolve the session from the request's access token.
		sess := Session(clientIp, AuthToken(c))
		if sess == nil {
			event.AuditWarn([]string{clientIp, "oauth2", "unknown client", action, status.Denied})
			Abort(c, http.StatusUnauthorized, i18n.ErrUnauthorized)
			return
		}

		// Reject tokens not bound to an active, registered user account
		// (e.g. client-credentials sessions without a user).
		user := sess.GetUser()
		if user == nil || user.IsUnknown() || user.IsDisabled() || !user.IsRegistered() {
			event.AuditWarn([]string{clientIp, "oauth2", "session %s", action, status.Denied}, sess.RefID)
			Abort(c, http.StatusUnauthorized, i18n.ErrUnauthorized)
			return
		}

		// Assemble the standard OpenID Connect UserInfo claims from the account.
		info := &oidc.UserInfo{
			Subject: user.UserUID,
			UserInfoProfile: oidc.UserInfoProfile{
				Name:              user.FullName(),
				PreferredUsername: user.Username(),
			},
			UserInfoEmail: oidc.UserInfoEmail{
				Email:         user.Email(),
				EmailVerified: oidc.Bool(user.VerifiedAt != nil),
			},
		}

		event.AuditInfo([]string{clientIp, "oauth2", "session %s", action, status.Granted}, sess.RefID)

		c.JSON(http.StatusOK, info)
	}

	router.GET("/oauth/userinfo", handler)
	router.POST("/oauth/userinfo", handler)
}
