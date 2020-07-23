package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/service"
)

// GET /api/v1/config
func GetConfig(router *gin.RouterGroup) {
	router.GET("/config", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceConfig, acl.ActionRead)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if s.User.Guest() {
			c.JSON(http.StatusOK, conf.GuestConfig())
		} else if s.User.Registered() {
			c.JSON(http.StatusOK, conf.UserConfig())
		} else {
			c.JSON(http.StatusOK, conf.PublicConfig())
		}
	})
}
