package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// Echo returns the request and response headers as JSON if debug mode is enabled.
//
//	@Summary	returns the request and response headers as JSON if debug mode is enabled
//	@Id			Echo
//	@Tags		Debug
//	@Success	200
//	@Router		/api/v1/echo [get]
func Echo(router *gin.RouterGroup) {
	methods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodTrace,
	}
	router.Match(methods, "/echo", func(c *gin.Context) {
		// Return if only headers are requested.
		if c.Request.Method == http.MethodHead {
			return
		}

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
