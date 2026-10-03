package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// revokeSession resolves the session a revoke request carries in its header, without charging
// the authentication failure budget. The handler owns that accounting, because a header that does
// not resolve is not a failed authentication on its own: the request may still prove possession
// with the token it submits, and then it is not charged at all.
func revokeSession(authToken string) *entity.Session {
	if !rnd.IsAuthAny(authToken) {
		return nil
	}

	sess, err := entity.FindSession(rnd.SessionID(authToken))

	if err != nil {
		return nil
	}

	return sess
}

// OAuthRevoke revokes an access token or session. A client may only revoke its own tokens.
//
//	@Summary	revoke an OAuth2 access token or session
//	@Id			OAuthRevoke
//	@Tags		Authentication
//	@Accept		json,x-www-form-urlencoded,mpfd
//	@Produce	json
//	@Param		request					body		form.OAuthRevokeToken	true	"revoke request"
//	@Success	200						{object}	gin.H
//	@Failure	400,401,403,404,413,429	{object}	i18n.Response
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
			event.AuditErr([]string{clientIp, "oauth2", "%s", action, authn.ErrDisabledInPublicMode.Error()}, actor)
			Abort(c, http.StatusForbidden, i18n.ErrForbidden)
			return
		}

		// Abort if the client has exhausted its authentication failure budget, before the
		// session lookup and the body are read. Reject only reads the budget; the charging
		// rule is set up below.
		if limiter.Auth.Reject(clientIp) {
			event.AuditWarn([]string{clientIp, "oauth2", "%s", action, "rate limit exceeded", status.Denied}, actor)
			limiter.AbortJSON(c)
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
		} else if s = revokeSession(authToken); s != nil {
			// Set log role and actor based on the session referenced in request header.
			sUserUID = s.UserUID
			if s.IsClient() {
				role = s.GetClientRole()
				actor = fmt.Sprintf("client %s", clean.LogQuote(s.GetClientInfo()))
			} else if username := s.GetUserName(); username != "" {
				role = s.GetUserRole()
				actor = fmt.Sprintf("user %s", clean.LogQuote(username))
			} else {
				role = s.GetUserRole()
				actor = fmt.Sprintf("unknown %s", s.GetUserRole().String())
			}
		}

		// Count this request against the authentication failure budget at most once, and only
		// while it has proved possession of nothing. The header lookup above does not charge,
		// so every charge on this route is made here; and once a session resolves, a refusal
		// below is an authorization outcome, which this budget does not meter.
		charged := false
		proven := s != nil

		charge := func() {
			if charged || proven {
				return
			}

			charged = true
			limiter.Auth.Reserve(clientIp)
		}

		LimitRequestBodyBytes(c, MaxOAuthRequestBytes)

		// Get the auth token to be revoked from the submitted form values or the request header.
		err = BindOAuthRequest(c, &frm)

		switch {
		case IsRequestBodyTooLarge(err):
			charge()
			event.AuditWarn([]string{clientIp, "oauth2", "%s", action, "request too large", status.Error(err)}, actor)
			AbortRequestTooLarge(c, i18n.ErrBadRequest)
			return
		case errors.Is(err, ErrUnsupportedContentType):
			// A body that cannot be decoded is an error, not an absent body: the
			// fallback below applies to a request that carries no body at all.
			charge()
			event.AuditWarn([]string{clientIp, "oauth2", "%s", action, status.Error(err)}, actor)
			AbortBadRequest(c, err)
			return
		case err != nil && authToken == "":
			charge()
			event.AuditWarn([]string{clientIp, "oauth2", "%s", action, status.Error(err)}, actor)
			AbortBadRequest(c, err)
			return
		case frm.Empty():
			frm.Token = authToken
			frm.TokenTypeHint = form.AccessToken
		}

		// Validate revocation form values.
		if err = frm.Validate(); err != nil {
			charge()
			event.AuditWarn([]string{clientIp, "oauth2", "%s", action, status.Error(err)}, actor)
			AbortInvalidCredentials(c)
			return
		}

		// Find session to be revoked.
		switch frm.TokenTypeHint {
		case form.RefID:
			if s == nil || sUserUID == "" || role == acl.RoleNone {
				charge()
				event.AuditWarn([]string{clientIp, "oauth2", "%s", action, "ref id requires a session", status.Denied}, actor)
				c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
				return
			} else if sess = entity.FindSessionByRefID(frm.Token); sess == nil {
				charge()
				event.AuditWarn([]string{clientIp, "oauth2", "%s", action, "ref id not found", status.Denied}, actor)
				AbortInvalidCredentials(c)
				return
			}
		case form.SessionID:
			if s == nil || sUserUID == "" || role == acl.RoleNone {
				charge()
				event.AuditWarn([]string{clientIp, "oauth2", "%s", action, "session id requires a session", status.Denied}, actor)
				c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
				return
			}

			sess, err = entity.FindSession(frm.Token)
		case form.AccessToken:
			sess, err = entity.FindSession(rnd.SessionID(frm.Token))
		}

		// A session that resolved from a submitted secret proves the caller holds it, so the
		// refusals below are authorization outcomes rather than failed authentications. A ref
		// id is an identifier rather than a secret, so it proves nothing on its own; that path
		// is reached only with a session the request header already proved.
		if err == nil && sess != nil && frm.TokenTypeHint != form.RefID {
			proven = true
		}

		// If not already set, get the log role and actor from the session to be revoked.
		if sess != nil && role == acl.RoleNone {
			if sess.IsClient() {
				role = sess.GetClientRole()
				actor = fmt.Sprintf("client %s", clean.LogQuote(sess.GetClientInfo()))
			} else if username := sess.GetUserName(); username != "" {
				role = s.GetUserRole()
				actor = fmt.Sprintf("user %s", clean.LogQuote(username))
			} else {
				role = sess.GetUserRole()
				actor = fmt.Sprintf("unknown %s", sess.GetUserRole().String())
			}
		}

		// Check revocation request and abort if invalid.
		switch {
		case err != nil:
			charge()
			event.AuditErr([]string{clientIp, "oauth2", "%s", action, "delete %s as %s", status.Error(err)}, actor, clean.Log(sess.RefID), role.String())
			AbortInvalidCredentials(c)
			return
		case sess == nil:
			charge()
			event.AuditErr([]string{clientIp, "oauth2", "%s", action, "delete %s as %s", status.Denied}, actor, "", role.String())
			AbortInvalidCredentials(c)
			return
		case sess.Abort(c):
			charge()
			event.AuditErr([]string{clientIp, "oauth2", "%s", action, "delete %s as %s", status.Denied}, actor, clean.Log(sess.RefID), role.String())
			return
		case !sess.IsClient():
			charge()
			event.AuditErr([]string{clientIp, "oauth2", "%s", action, "delete %s as %s", status.Denied}, actor, clean.Log(sess.RefID), role.String())
			c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
			return
		case sUserUID != "" && sess.UserUID != sUserUID:
			charge()
			event.AuditErr([]string{clientIp, "oauth2", "%s", action, "delete %s as %s", authn.ErrUnauthorized.Error()}, actor, clean.Log(sess.RefID), role.String())
			AbortInvalidCredentials(c)
			return
		default:
			event.AuditInfo([]string{clientIp, "oauth2", "%s", action, "delete %s as %s", status.Granted}, actor, clean.Log(sess.RefID), role.String())
		}

		// Delete session cache and database record.
		if err = sess.Delete(); err != nil {
			// Log error.
			event.AuditErr([]string{clientIp, "oauth2", "%s", action, "delete %s as %s", status.Error(err)}, actor, clean.Log(sess.RefID), role.String())

			// Return JSON error.
			c.AbortWithStatusJSON(http.StatusNotFound, i18n.NewResponse(http.StatusNotFound, i18n.ErrNotFound))
			return
		}

		// Log event.
		event.AuditInfo([]string{clientIp, "oauth2", "%s", action, "delete %s as %s", "deleted"}, actor, clean.Log(sess.RefID), role.String())

		// Send response.
		c.JSON(http.StatusOK, DeleteSessionResponse(sess.ID))
	})
}
