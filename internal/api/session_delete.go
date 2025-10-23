package api

import (
	"net/http"

	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/log/status"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/session"
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

		// Check if the session user is allowed to manage all accounts or update his/her own account.
		s := AuthAny(c, acl.ResourceSessions, acl.Permissions{acl.ActionManage, acl.ActionDelete})

		if s.Abort(c) {
			return
		}

		id := clean.ID(c.Param("id"))

		// Get client IP and auth token from request headers.
		clientIp := ClientIP(c)

		// Only admins may delete other sessions by ref id.
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

		// Return JSON response for confirmation.
		c.JSON(http.StatusOK, DeleteSessionResponse(s.ID))
	}

	router.DELETE("/session", deleteSessionHandler)
	router.DELETE("/session/:id", deleteSessionHandler)
	router.DELETE("/sessions/:id", deleteSessionHandler)
}
