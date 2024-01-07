package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// DeleteSession deletes an existing client session (logout).
//
// DELETE /api/v1/session/:id
func DeleteSession(router *gin.RouterGroup) {
	router.DELETE("/session/:id", func(c *gin.Context) {
		id := clean.ID(c.Param("id"))

		// Abort if authentication token is missing or empty.
		if id == "" {
			AbortBadRequest(c)
			return
		} else if get.Config().Public() {
			// Return JSON response for confirmation.
			c.JSON(http.StatusOK, DeleteSessionResponse(id))
			return
		}

		// Only admins may delete other sessions by reference id.
		if rnd.IsRefID(id) {
			if s := Session(AuthToken(c)); s == nil {
				entity.SessionStatusUnauthorized().Abort(c)
				return
			} else if s.Abort(c) {
				return
			} else if !acl.Resources.AllowAll(acl.ResourceUsers, s.User().AclRole(), acl.Permissions{acl.AccessAll, acl.ActionManage}) {
				s.Abort(c)
				return
			} else if ref := entity.FindSessionByRefID(id); ref == nil {
				AbortNotFound(c)
				return
			} else {
				id = ref.ID
			}
		} else {
			if s := Session(AuthToken(c)); s == nil {
				entity.SessionStatusUnauthorized().Abort(c)
				return
			} else if s.Abort(c) {
				return
			} else if s.ID != id {
				entity.SessionStatusForbidden().Abort(c)
				return
			}
		}

		// Delete session cache and database record.
		if err := get.Session().Delete(id); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s"}, err)
		} else {
			event.AuditDebug([]string{ClientIP(c), "session deleted"})
		}

		// Return JSON response for confirmation.
		c.JSON(http.StatusOK, DeleteSessionResponse(id))
	})
}
