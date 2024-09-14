package api

import (
	"net/http"
	"os"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// StopServer shuts down the server.
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

		process, err := os.FindProcess(os.Getpid())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, NewResponse(http.StatusInternalServerError, err, ""))
			return
		} else {
			c.JSON(http.StatusOK, conf.Options())
		}

		if err = process.Signal(syscall.SIGTERM); err != nil {
			log.Errorf("server: %s", err)
		}
	})
}
