package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetStatus reports if the server is operational.
//
//	@Summary	reports if the server is operational
//	@Id			GetStatus
//	@Tags		Server
//	@Produce	json
//	@Success	200	{object}	gin.H
//	@Router		/api/v1/status [get]
func GetStatus(router *gin.RouterGroup) {
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "operational"})
	})
}
