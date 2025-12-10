package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// OAuthRevoke revokes an access token or session. A client may only revoke its own tokens.
//
//	@Summary	revoke an OAuth2 access token or session
//	@Id			OAuthRevoke
//	@Tags		Authentication
//	@Accept		json
//	@Produce	json
//	@Param		request				body		form.OAuthRevokeToken	true	"revoke request"
//	@Success	200					{object}	gin.H
//	@Failure	400,401,403,404,429	{object}	i18n.Response
//	@Router		/api/v1/oauth/revoke [post]
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

		// Token revocation request form.
		var frm form.OAuthRevokeToken

		// Get token and session from request header.
		if authToken = AuthToken(c); authToken == "" {
			role = acl.RoleNone
		} else if s = Session(clientIp, authToken); s != nil {
			// Set log role and actor based on the session referenced in request header.
			sUserUID = s.UserUID
			if s.IsClient() {
				role = s.GetClientRole()
				actor = fmt.Sprintf("client %s", clean.Log(s.GetClientInfo()))
			} else if username := s.GetUserName(); username != "" {
				role = s.GetUserRole()
				actor = fmt.Sprintf("user %s", clean.Log(username))
			} else {
				role = s.GetUserRole()
				actor = fmt.Sprintf("unknown %s", s.GetUserRole().String())
			}
		}

		// Get the auth token to be revoked from the submitted form values or the request header.
		if err = c.ShouldBind(&frm); err != nil && authToken == "" {
			event.AuditWarn([]string{clientIp, "oauth2", actor, action, status.Error(err)})
			AbortBadRequest(c, err)
			return
		} else if frm.Empty() {
			frm.Token = authToken
			frm.TokenTypeHint = form.AccessToken
		}

		// Validate revocation form values.
		if err = frm.Validate(); err != nil {
			event.AuditWarn([]string{clientIp, "oauth2", actor, action, status.Error(err)})
			AbortInvalidCredentials(c)
			return
		}

		// Find session to be revoked.
		switch frm.TokenTypeHint {
		case form.RefID:
			if s == nil || sUserUID == "" || role == acl.RoleNone {
				c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
				return
			} else if sess = entity.FindSessionByRefID(frm.Token); sess == nil {
				AbortInvalidCredentials(c)
				return
			}
		case form.SessionID:
			if s == nil || sUserUID == "" || role == acl.RoleNone {
				c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
				return
			}

			sess, err = entity.FindSession(frm.Token)
		case form.AccessToken:
			sess, err = entity.FindSession(rnd.SessionID(frm.Token))
		}

		// If not already set, get the log role and actor from the session to be revoked.
		if sess != nil && role == acl.RoleNone {
			if sess.IsClient() {
				role = sess.GetClientRole()
				actor = fmt.Sprintf("client %s", clean.Log(sess.GetClientInfo()))
			} else if username := sess.GetUserName(); username != "" {
				role = s.GetUserRole()
				actor = fmt.Sprintf("user %s", clean.Log(username))
			} else {
				role = sess.GetUserRole()
				actor = fmt.Sprintf("unknown %s", sess.GetUserRole().String())
			}
		}

		// Check revocation request and abort if invalid.
		switch {
		case err != nil:
			event.AuditErr([]string{clientIp, "oauth2", actor, action, "delete %s as %s", status.Error(err)}, clean.Log(sess.RefID), role.String())
			AbortInvalidCredentials(c)
			return
		case sess == nil:
			event.AuditErr([]string{clientIp, "oauth2", actor, action, "delete %s as %s", status.Denied}, "", role.String())
			AbortInvalidCredentials(c)
			return
		case sess.Abort(c):
			event.AuditErr([]string{clientIp, "oauth2", actor, action, "delete %s as %s", status.Denied}, clean.Log(sess.RefID), role.String())
			return
		case !sess.IsClient():
			event.AuditErr([]string{clientIp, "oauth2", actor, action, "delete %s as %s", status.Denied}, clean.Log(sess.RefID), role.String())
			c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
			return
		case sUserUID != "" && sess.UserUID != sUserUID:
			event.AuditErr([]string{clientIp, "oauth2", actor, action, "delete %s as %s", authn.ErrUnauthorized.Error()}, clean.Log(sess.RefID), role.String())
			AbortInvalidCredentials(c)
			return
		default:
			event.AuditInfo([]string{clientIp, "oauth2", actor, action, "delete %s as %s", status.Granted}, clean.Log(sess.RefID), role.String())
		}

		// Delete session cache and database record.
		if err = sess.Delete(); err != nil {
			// Log error.
			event.AuditErr([]string{clientIp, "oauth2", actor, action, "delete %s as %s", status.Error(err)}, clean.Log(sess.RefID), role.String())

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
