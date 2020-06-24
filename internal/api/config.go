package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
)

// GET /api/v1/config
func GetConfig(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/config", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		sess := Session(SessionToken(c), conf)

		if sess == nil {
			c.JSON(http.StatusNotFound, ErrSessionNotFound)
			return
		}

		if sess.User.Guest() {
			c.JSON(http.StatusOK, conf.GuestConfig())
		} else if sess.User.User() {
			c.JSON(http.StatusOK, conf.UserConfig())
		} else {
			c.JSON(http.StatusOK, conf.PublicConfig())
		}
	})
}
