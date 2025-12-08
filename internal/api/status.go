package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetStatus responds with status code 200 if the server is operational.
//
//	@Summary	responds with status code 200 if the server is operational
//	@Id			GetStatus
//	@Tags		Debug
//	@Produce	json
//	@Success	200	{object}	gin.H
//	@Router		/api/v1/status [get]
func GetStatus(router *gin.RouterGroup) {
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "operational"})
	})
}
