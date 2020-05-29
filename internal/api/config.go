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

		c.JSON(http.StatusOK, conf.ClientConfig())
	})
}
