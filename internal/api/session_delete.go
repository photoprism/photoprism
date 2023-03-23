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

		// Abort if ID is missing.
		if id == "" {
			AbortBadRequest(c)
			return
		} else if get.Config().Public() {
			c.JSON(http.StatusOK, gin.H{"status": "authentication disabled", "id": id})
			return
		}

		// Find session by reference ID.
		if !rnd.IsRefID(id) {
			// Do nothing.
		} else if s := Session(SessionID(c)); s == nil {
			entity.SessionStatusUnauthorized().Abort(c)
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

		// Delete session by ID.
		if err := get.Session().Delete(id); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s"}, err)
		} else {
			event.AuditDebug([]string{ClientIP(c), "session deleted"})
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "id": id})
	})
}
