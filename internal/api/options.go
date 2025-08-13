package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Options returns an empty response to handle CORS preflight requests.
//
//	@Summary	returns CORS headers with an empty response body
//	@Id			Options
//	@Success	204
//	@Router		/api/v1/{any} [options]
func Options(router *gin.RouterGroup) {
	router.OPTIONS("/*any", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}
