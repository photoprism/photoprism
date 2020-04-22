package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
)

// GET /api/v1/status
func GetStatus(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "operational"})
	})
}
