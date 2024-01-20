package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/header"
)

// Api is a middleware that sets additional response headers when serving REST API requests.
var Api = func(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set vary response header.
		c.Header(header.Vary, header.DefaultVary)

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
	}
}
