package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/v1/status
func GetStatus(router *gin.RouterGroup) {
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "operational"})
	})
}
