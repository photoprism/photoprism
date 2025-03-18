package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/process"
)

// StopServer initiates a server restart if the user is authorized.
//
// POST /api/v1/server/stop
func StopServer(router *gin.RouterGroup) {
	router.POST("/server/stop", func(c *gin.Context) {
		s := Auth(c, acl.ResourceConfig, acl.ActionManage)
		conf := get.Config()

		// Abort if permission was not granted.
		if s.Invalid() || conf.Public() || conf.DisableSettings() || conf.DisableRestart() {
			AbortForbidden(c)
			return
		}

		// Trigger restart.
		//
		// Note that this requires an entrypoint script or other process to
		// spawns a new instance when the server exists with status code 1.
		c.JSON(http.StatusOK, conf.Options())
		process.Restart()
	})
}
