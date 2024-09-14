package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// Echo returns the request and response headers as JSON if debug mode is enabled.
//
// The supported request methods are:
//
//   - GET
//   - POST
//   - PUT
//   - PATCH
//   - HEAD
//   - OPTIONS
//   - DELETE
//   - CONNECT
//   - TRACE
//
// ANY /api/v1/echo
func Echo(router *gin.RouterGroup) {
	router.Any("/echo", func(c *gin.Context) {
		// Abort if debug mode is disabled.
		if !get.Config().Debug() {
			AbortFeatureDisabled(c)
			return
		} else if c.Request == nil || c.Writer == nil {
			AbortUnexpectedError(c)
			return
		}

		// Return request information.
		echoResponse := gin.H{
			"url":    c.Request.URL.String(),
			"method": c.Request.Method,
			"headers": map[string]http.Header{
				"request":  c.Request.Header,
				"response": c.Writer.Header(),
			},
		}

		c.JSON(http.StatusOK, echoResponse)
	})
}
