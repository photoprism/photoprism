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
		// Only allow crawlers to index the site if it is a public demo (or if there is a public image wall):
		// https://github.com/photoprism/photoprism/issues/669
		if !conf.Demo() || !conf.Public() {
			// Set "X-Robots-Tag" header:
			// https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#xrobotstag
			c.Header(header.RobotsTag, header.RobotsNone)
		}

		// Abort if the request must not be served through a CDN.
		if header.AbortCdnRequest(c.Request) {
			api.AbortNotFound(c)
			return
		}

		// Set "Content-Security-Policy" header:
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
		c.Header(header.ContentSecurityPolicy, header.DefaultContentSecurityPolicy)

		// Set "X-Frame-Options" header:
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options
		c.Header(header.FrameOptions, header.DefaultFrameOptions)
	}
}
