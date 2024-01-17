package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/header"
)

// Static is a middleware that adds static content-related headers to the server's response.
var Static = func(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow caching of static assets for up to 7 days.
		header.SetCacheControlImmutable(c, 0, true)

		// Allow CORS based on the configuration or otherwise automatically for certain file types when using a CDN.
		// See: https://www.w3.org/TR/css-fonts-3/#font-fetching-requirements
		if origin := conf.CORSOrigin(); origin != "" || header.AllowCORS(c.Request.URL.Path) && conf.UseCdn() {
			if origin == "" {
				c.Header(header.AccessControlAllowOrigin, header.Any)
			} else {
				c.Header(header.AccessControlAllowOrigin, origin)
			}

			// Add additional information to preflight OPTION requests.
			if c.Request.Method == http.MethodOptions {
				c.Header(header.AccessControlAllowHeaders, conf.CORSHeaders())
				c.Header(header.AccessControlAllowMethods, conf.CORSMethods())
				c.Header(header.AccessControlMaxAge, header.DefaultAccessControlMaxAge)
			}
		}
	}
}
