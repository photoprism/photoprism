package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
)

// GET /api/v1/settings
func GetSettings(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/settings", func(c *gin.Context) {
		result := conf.Settings()

		c.JSON(http.StatusOK, result)
	})
}

// POST /api/v1/settings
func SaveSettings(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/settings", func(c *gin.Context) {
		// TODO

		c.JSON(http.StatusOK, gin.H{"message": "saved"})
	})
}
