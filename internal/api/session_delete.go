package api

import (
	"net/http"
	"strings"

	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/log/status"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/oidc"
	"github.com/photoprism/photoprism/internal/auth/session"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// DeleteSession deletes an existing client session (logout).
//
//	@Summary	delete a session (logout)
//	@Tags		Authentication
//	@Produce	json
//	@Param		id				path		string	false	"session id or ref id"
//	@Success	200				{object}	gin.H
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Router		/api/v1/session [delete]
//	@Router		/api/v1/session/{id} [delete]
//	@Router		/api/v1/sessions/{id} [delete]
func DeleteSession(router *gin.RouterGroup) {
	deleteSessionHandler := func(c *gin.Context) {
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			AbortNotFound(c)
			return
		}

		// Abort if running in public mode.
		if get.Config().Public() {
			c.JSON(http.StatusOK, DeleteSessionResponse(session.PublicID))
			return
		}

		// Require session management or delete access.
		s := AuthAny(c, acl.ResourceSessions, acl.Permissions{acl.ActionManage, acl.ActionDelete})

		if s.Abort(c) {
			return
		}

		id := clean.ID(c.Param("id"))

		// Get client IP and auth token from request headers.
		clientIp := ClientIP(c)

		// Only full session managers may delete sessions by ref ID.
		if rnd.IsRefID(id) {
			if !acl.Rules.AllowAll(acl.ResourceSessions, s.GetUserRole(), acl.Permissions{acl.AccessAll, acl.ActionManage}) {
				event.AuditErr([]string{clientIp, "session %s", "delete %s as %s", status.Denied}, s.RefID, acl.ResourceSessions.String(), s.GetUserRole())
				Abort(c, http.StatusForbidden, i18n.ErrForbidden)
				return
			}

			event.AuditInfo([]string{clientIp, "session %s", "delete %s as %s", status.Granted}, s.RefID, acl.ResourceSessions.String(), s.GetUserRole())

			if s = entity.FindSessionByRefID(id); s == nil {
				Abort(c, http.StatusNotFound, i18n.ErrNotFound)
				return
			}
		} else if id != "" && s.ID != id {
			event.AuditWarn([]string{clientIp, "session %s", "delete %s as %s", "ids do not match"}, s.RefID, acl.ResourceSessions.String(), s.GetUserRole())
			Abort(c, http.StatusForbidden, i18n.ErrForbidden)
			return
		}

		// Delete session cache and database record.
		if err := s.Delete(); err != nil {
			event.AuditErr([]string{clientIp, "session %s", "delete session as %s", status.Error(err)}, s.RefID, s.GetUserRole())
		} else {
			event.AuditDebug([]string{clientIp, "session %s", "deleted"}, s.RefID)
		}

		resp := DeleteSessionResponse(s.ID)

		// On the caller's own logout (not a manager deleting by ref id), end the provider
		// session too when PHOTOPRISM_OIDC_LOGOUT is enabled.
		if conf := get.Config(); !rnd.IsRefID(id) {
			if logoutUri := oidcLogoutURL(conf, get.OIDC(), s); logoutUri != "" {
				// Delegate to the provider's end-session endpoint (the Portal OP in a cluster),
				// which clears the OP session cookie itself — clearing it here would strip the
				// cookie the Portal OP needs to resolve the session and chain the upstream logout.
				resp["providerLogoutUri"] = logoutUri
				event.AuditInfo([]string{clientIp, "session %s", "oidc provider logout initiated", status.Granted}, s.RefID)
			} else if clearPath := OIDCSessionCookieClearPath(conf); clearPath != "" {
				// Otherwise clear the OP cookie locally so cluster-wide Sign-Out stops silent re-SSO.
				ClearOIDCSessionCookie(c, clearPath, conf.SiteHttps())
			}
		}

		// Return JSON response for confirmation.
		c.JSON(http.StatusOK, resp)
	}

	router.DELETE("/session", deleteSessionHandler)
	router.DELETE("/session/:id", deleteSessionHandler)
	router.DELETE("/sessions/:id", deleteSessionHandler)
}

// oidcLogoutURL returns the RP-initiated logout URL for a just-deleted OIDC session, or ""
// when logout is disabled, the session was not OIDC, or the provider advertises no
// end_session_endpoint. The browser is redirected there to end the provider SSO session.
func oidcLogoutURL(conf *config.Config, provider *oidc.Client, s *entity.Session) string {
	if conf == nil || provider == nil || s == nil {
		return ""
	} else if !conf.OIDCLogout() || s.IdToken == "" || !s.GetProvider().IsOIDC() {
		return ""
	}

	logoutUri, err := provider.EndSessionURL(s.IdToken, AbsoluteLoginURL(conf), "")

	if err != nil {
		event.AuditWarn([]string{"oidc", "provider logout", status.Error(err)})
		return ""
	}

	return logoutUri
}

// AbsoluteLoginURL returns the node's login page as an absolute URL, suitable as an OIDC
// post_logout_redirect_uri (providers require it absolute). LoginUri() is root-relative,
// so it is joined to the site origin.
func AbsoluteLoginURL(conf *config.Config) string {
	path := conf.LoginUri()

	if strings.Contains(path, "://") {
		return path
	}

	origin := scheme.OriginURL(conf.SiteUrl())

	if origin == "" {
		return path
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return strings.TrimRight(origin, "/") + path
}
