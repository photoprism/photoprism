package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Options returns CORS headers with an empty response body.
//
//	@Summary		returns CORS headers with an empty response body
//	@Description	A preflight request is automatically issued by a browser and in normal cases, front-end developers don't need to craft such requests themselves. It appears when request is qualified as "to be preflighted" and omitted for simple requests.
//	@Id				Options
//	@Tags			CORS
//	@Success		204
//	@Router			/api/v1/{any} [options]
func Options(router *gin.RouterGroup) {
	router.OPTIONS("/*any", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
}
