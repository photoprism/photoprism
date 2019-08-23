package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
)

// GET /api/v1/ping
func Ping(router *gin.RouterGroup, _ *config.Config) {
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})
}
