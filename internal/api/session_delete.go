package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/session"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// DeleteSession deletes an existing client session (logout).
//
// DELETE /api/v1/session
// DELETE /api/v1/session/:id
func DeleteSession(router *gin.RouterGroup) {
	deleteSessionHandler := func(c *gin.Context) {
		// Abort if running in public mode.
		if get.Config().Public() {
			// Return JSON response for confirmation.
			c.JSON(http.StatusOK, DeleteSessionResponse(session.PublicID))
			return
		}

		id := clean.ID(c.Param("id"))

		// Get client IP address for logs and rate limiting checks.
		clientIP := ClientIP(c)

		// Find session based on auth token.
		sess, err := entity.FindSession(rnd.SessionID(AuthToken(c)))

		if err != nil || sess == nil {
			Abort(c, http.StatusUnauthorized, i18n.ErrUnauthorized)
			return
		} else if sess.Abort(c) {
			return
		}

		// Only admins may delete other sessions by ref id.
		if rnd.IsRefID(id) {
			if !acl.Resources.AllowAll(acl.ResourceUsers, sess.User().AclRole(), acl.Permissions{acl.AccessAll, acl.ActionManage}) {
				event.AuditErr([]string{clientIP, "session %s", "delete session %s as %s", "denied"}, sess.RefID, id, sess.User().AclRole())
				Abort(c, http.StatusForbidden, i18n.ErrForbidden)
				return
			}

			event.AuditInfo([]string{clientIP, "session %s", "delete session %s as %s", "granted"}, sess.RefID, id, sess.User().AclRole())

			if sess = entity.FindSessionByRefID(id); sess == nil {
				Abort(c, http.StatusNotFound, i18n.ErrNotFound)
				return
			}
		} else if id != "" && sess.ID != id {
			event.AuditWarn([]string{clientIP, "session %s", "delete session as %s", "ids do not match"}, sess.RefID, sess.User().AclRole())
			Abort(c, http.StatusForbidden, i18n.ErrForbidden)
			return
		}

		// Delete session cache and database record.
		if err = sess.Delete(); err != nil {
			event.AuditErr([]string{clientIP, "session %s", "delete session as %s", "%s"}, sess.RefID, sess.User().AclRole(), err)
		} else {
			event.AuditDebug([]string{clientIP, "session %s", "deleted"}, sess.RefID)
		}

		// Return JSON response for confirmation.
		c.JSON(http.StatusOK, DeleteSessionResponse(sess.ID))
	}

	router.DELETE("/session", deleteSessionHandler)
	router.DELETE("/session/:id", deleteSessionHandler)
}
