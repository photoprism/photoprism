package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
)

// DeleteSession deletes an existing client session (logout).
//
// DELETE /api/v1/session/:id
func DeleteSession(router *gin.RouterGroup) {
	router.DELETE("/session/:id", func(c *gin.Context) {
		id := clean.ID(c.Param("id"))

		if id == "" {
			AbortBadRequest(c)
			return
		} else if service.Config().Public() {
			c.JSON(http.StatusOK, gin.H{"status": "authentication disabled", "id": id})
			return
		}

		if err := service.Session().Delete(id); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s"}, err)
		} else {
			event.AuditDebug([]string{ClientIP(c), "session deleted"})
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "id": id})
	})
}
