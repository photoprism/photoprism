package server

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/header"
)

// Robots is a middleware that adds a "X-Robots-Tag" header to the response:
// https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#xrobotstag
var Robots = func(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Block search engines until a public picture wall has been implemented,
		// see https://github.com/photoprism/photoprism/issues/669.
		c.Header(header.Robots, header.RobotsNone)
	}
}
