package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/process"
)

// StopServer allows authorized admins to restart the server.
//
//	@Summary	allows authorized admins to restart the server
//	@Id			StopServer
//	@Tags		Internal
//	@Produce	json
//	@Success	200		{object}	config.Options
//	@Failure	401,403	{object}	i18n.Response
//	@Router		/api/v1/server/stop [post]
func StopServer(router *gin.RouterGroup) {
	router.POST("/server/stop", func(c *gin.Context) {
		s := Auth(c, acl.ResourceConfig, acl.ActionManage)

		conf := get.Config()

		// Abort if permission is not granted.
		if s.Invalid() || conf.Public() || conf.DisableSettings() || conf.DisableRestart() {
			AbortForbidden(c)
			return
		}

		options := conf.Options()

		// Trigger restart.
		//
		// Note that this requires an entrypoint script or other process to
		// spawns a new instance when the server exists with status code 1.
		c.JSON(http.StatusOK, options)
		process.Restart()
	})
}
