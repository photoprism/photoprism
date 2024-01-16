package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/header"
)

// Security adds common HTTP security headers to the response.
var Security = func(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only allow CDNs to cache responses of GET, HEAD, and OPTIONS requests and block the request otherwise.
		if header.BlockCdn(c.Request) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// If permitted, set CORS headers (Cross-Origin Resource Sharing).
		// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
		if origin := conf.CORSOrigin(); origin != "" {
			c.Header(header.AccessControlAllowOrigin, origin)

			// Add additional information to preflight OPTION requests.
			if c.Request.Method == http.MethodOptions {
				c.Header(header.AccessControlAllowHeaders, conf.CORSHeaders())
				c.Header(header.AccessControlAllowMethods, conf.CORSMethods())
				c.Header(header.AccessControlMaxAge, header.DefaultAccessControlMaxAge)
			}
		}

		// Set Content Security Policy.
		// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
		c.Header(header.ContentSecurityPolicy, header.DefaultContentSecurityPolicy)

		// Set Frame Options.
		// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options
		c.Header(header.FrameOptions, header.DefaultFrameOptions)
	}
}
