package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
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

// OAuthRevoke takes an access token and deletes it. A client may only delete its own tokens.
//
// POST /api/v1/oauth/revoke
func OAuthRevoke(router *gin.RouterGroup) {
	router.POST("/oauth/revoke", func(c *gin.Context) {
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
		action := "revoke token"

		// Abort if running in public mode.
		if get.Config().Public() {
			event.AuditErr([]string{clientIp, "oauth2", actor, action, authn.ErrDisabledInPublicMode.Error()})
			Abort(c, http.StatusForbidden, i18n.ErrForbidden)
			return
		}

		// Session and user information.
		var s, sess *entity.Session
		var authToken, sUserUID string
		var role acl.Role
		var err error

		// Token revokation request form.
		var f form.OAuthRevokeToken

		// Get token and session from request header.
		if authToken = AuthToken(c); authToken == "" {
			role = acl.RoleNone
		} else if s = Session(clientIp, authToken); s != nil {
			// Set log role and actor based on the session referenced in request header.
			sUserUID = s.UserUID
			if s.IsClient() {
				role = s.ClientRole()
				actor = fmt.Sprintf("client %s", clean.Log(s.ClientInfo()))
			} else if username := s.Username(); username != "" {
				role = s.UserRole()
				actor = fmt.Sprintf("user %s", clean.Log(username))
			} else {
				role = s.UserRole()
				actor = fmt.Sprintf("unknown %s", s.UserRole().String())
			}
		}

		// Get the auth token to be revoked from the submitted form values or the request header.
		if err = c.ShouldBind(&f); err != nil && authToken == "" {
			event.AuditWarn([]string{clientIp, "oauth2", actor, action, "%s"}, err)
			AbortBadRequest(c)
			return
		} else if f.Empty() {
			f.Token = authToken
			f.TokenTypeHint = form.AccessToken
		}

		// Validate revokation form values.
		if err = f.Validate(); err != nil {
			event.AuditWarn([]string{clientIp, "oauth2", actor, action, "%s"}, err)
			AbortInvalidCredentials(c)
			return
		}

		// Find session to be revoked.
		switch f.TokenTypeHint {
		case form.RefID:
			if s == nil || sUserUID == "" || role == acl.RoleNone {
				c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
				return
			} else if sess = entity.FindSessionByRefID(f.Token); sess == nil {
				AbortInvalidCredentials(c)
				return
			}
		case form.SessionID:
			if s == nil || sUserUID == "" || role == acl.RoleNone {
				c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
				return
			}

			sess, err = entity.FindSession(f.Token)
		case form.AccessToken:
			sess, err = entity.FindSession(rnd.SessionID(f.Token))
		}

		// If not already set, get the log role and actor from the session to be revoked.
		if sess != nil && role == acl.RoleNone {
			if sess.IsClient() {
				role = sess.ClientRole()
				actor = fmt.Sprintf("client %s", clean.Log(sess.ClientInfo()))
			} else if username := sess.Username(); username != "" {
				role = s.UserRole()
				actor = fmt.Sprintf("user %s", clean.Log(username))
			} else {
				role = sess.UserRole()
				actor = fmt.Sprintf("unknown %s", sess.UserRole().String())
			}
		}

		// Check revokation request and abort if invalid.
		if err != nil {
			event.AuditErr([]string{clientIp, "oauth2", actor, action, "delete %s as %s", "%s"}, clean.Log(sess.RefID), role.String(), err.Error())
			AbortInvalidCredentials(c)
			return
		} else if sess == nil {
			event.AuditErr([]string{clientIp, "oauth2", actor, action, "delete %s as %s", authn.Denied}, clean.Log(sess.RefID), role.String())
			AbortInvalidCredentials(c)
			return
		} else if sess.Abort(c) {
			event.AuditErr([]string{clientIp, "oauth2", actor, action, "delete %s as %s", authn.Denied}, clean.Log(sess.RefID), role.String())
			return
		} else if !sess.IsClient() {
			event.AuditErr([]string{clientIp, "oauth2", actor, action, "delete %s as %s", authn.Denied}, clean.Log(sess.RefID), role.String())
			c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
			return
		} else if sUserUID != "" && sess.UserUID != sUserUID {
			event.AuditErr([]string{clientIp, "oauth2", actor, action, "delete %s as %s", authn.ErrUnauthorized.Error()}, clean.Log(sess.RefID), role.String())
			AbortInvalidCredentials(c)
			return
		} else {
			event.AuditInfo([]string{clientIp, "oauth2", actor, action, "delete %s as %s", authn.Granted}, clean.Log(sess.RefID), role.String())
		}

		// Delete session cache and database record.
		if err = sess.Delete(); err != nil {
			// Log error.
			event.AuditErr([]string{clientIp, "oauth2", actor, action, "delete %s as %s", "%s"}, clean.Log(sess.RefID), role.String(), err)

			// Return JSON error.
			c.AbortWithStatusJSON(http.StatusNotFound, i18n.NewResponse(http.StatusNotFound, i18n.ErrNotFound))
			return
		}

		// Log event.
		event.AuditInfo([]string{clientIp, "oauth2", actor, action, "delete %s as %s", "deleted"}, clean.Log(sess.RefID), role.String())

		// Send response.
		c.JSON(http.StatusOK, DeleteSessionResponse(sess.ID))
	})
}
