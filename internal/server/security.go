package server

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/header"
)

// Security is a middleware that adds security-related headers to the server's response.
var Security = func(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Abort if the request should not be served through a CDN.
		if header.AbortCdnRequest(c.Request) {
			api.AbortNotFound(c)
			return
		}

		// Set Content Security Policy.
		// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
		c.Header(header.ContentSecurityPolicy, header.DefaultContentSecurityPolicy)

		// Set Frame Options.
		// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options
		c.Header(header.FrameOptions, header.DefaultFrameOptions)
	}
}
