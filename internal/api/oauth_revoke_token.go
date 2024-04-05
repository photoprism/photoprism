package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// RevokeOAuthToken takes an access token and deletes it. A client may only delete its own tokens.
//
// POST /api/v1/oauth/revoke
func RevokeOAuthToken(router *gin.RouterGroup) {
	router.POST("/oauth/revoke", func(c *gin.Context) {
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			AbortNotFound(c)
			return
		}

		// Get client IP address for logs and rate limiting checks.
		clientIp := ClientIP(c)

		// Abort if running in public mode.
		if get.Config().Public() {
			event.AuditErr([]string{clientIp, "client", "delete session", "oauth2", authn.ErrDisabledInPublicMode.Error()})
			Abort(c, http.StatusForbidden, i18n.ErrForbidden)
			return
		}

		var err error

		// Token revokation request form.
		var f form.OAuthRevokeToken

		// Get token from request header.
		authToken := AuthToken(c)

		// Get the auth token to be revoked from the submitted form values or the request header.
		if err = c.ShouldBind(&f); err != nil && authToken == "" {
			event.AuditWarn([]string{clientIp, "client", "delete session", "oauth2", "%s"}, err)
			AbortBadRequest(c)
			return
		} else if f.Empty() {
			f.AuthToken = authToken
			f.TypeHint = form.ClientAccessToken
		}

		// Check the token form values.
		if err = f.Validate(); err != nil {
			event.AuditWarn([]string{clientIp, "client", "delete session", "oauth2", "%s"}, err)
			AbortBadRequest(c)
			return
		}

		// Disable caching of responses.
		c.Header(header.CacheControl, header.CacheControlNoStore)

		// Find session based on auth token.
		sess, err := entity.FindSession(rnd.SessionID(f.AuthToken))

		if err != nil {
			event.AuditErr([]string{clientIp, "client %s", "session %s", "delete session as %s", "oauth2", "%s"}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID), sess.ClientRole().String(), err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, i18n.NewResponse(http.StatusUnauthorized, i18n.ErrUnauthorized))
			return
		} else if sess == nil {
			event.AuditErr([]string{clientIp, "client %s", "session %s", "delete session as %s", "oauth2", authn.Denied}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID), sess.ClientRole().String())
			c.AbortWithStatusJSON(http.StatusUnauthorized, i18n.NewResponse(http.StatusUnauthorized, i18n.ErrUnauthorized))
			return
		} else if sess.Abort(c) {
			event.AuditErr([]string{clientIp, "client %s", "session %s", "delete session as %s", "oauth2", authn.Denied}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID), sess.ClientRole().String())
			return
		} else if !sess.IsClient() {
			event.AuditErr([]string{clientIp, "client %s", "session %s", "delete session as %s", "oauth2", authn.Denied}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID), sess.ClientRole().String())
			c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
			return
		} else {
			event.AuditInfo([]string{clientIp, "client %s", "session %s", "delete session as %s", "oauth2", authn.Granted}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID), sess.ClientRole().String())
		}

		// Delete session cache and database record.
		if err = sess.Delete(); err != nil {
			// Log error.
			event.AuditErr([]string{clientIp, "client %s", "session %s", "delete session as %s", "oauth2", "%s"}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID), sess.ClientRole().String(), err)

			// Return JSON error.
			c.AbortWithStatusJSON(http.StatusNotFound, i18n.NewResponse(http.StatusNotFound, i18n.ErrNotFound))
			return
		}

		// Log event.
		event.AuditInfo([]string{clientIp, "client %s", "session %s", "oauth2", "deleted"}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID))

		// Send response.
		c.JSON(http.StatusOK, DeleteSessionResponse(sess.ID))
	})
}
