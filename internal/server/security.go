package server

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/header"
)

// Security adds common HTTP security headers to the response.
var Security = func(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow Cross-Origin Resource Sharing (CORS)?
		// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin#cors_and_caching
		if conf.HttpCORS() {
			c.Header(header.AccessControlAllowOrigin, header.Any)
		} else if origin := c.GetHeader(header.Origin); origin != "" {
			// Automatically set the "Access-Control-Allow-Origin" response header
			// based on the "Origin" value of the request.
			c.Header(header.AccessControlAllowOrigin, clean.Header(origin))
			c.Header(header.Vary, header.Origin)
		}

		// Set Content Security Policy.
		// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
		c.Header(header.ContentSecurityPolicy, header.DefaultContentSecurityPolicy)

		// Set Frame Options.
		// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options
		c.Header(header.FrameOptions, header.DefaultFrameOptions)
	}
}
