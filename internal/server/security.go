package server

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/header"
)

// Security adds common HTTP security headers to the response.
var Security = func(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set(header.ContentSecurityPolicy, header.DefaultContentSecurityPolicy)
		c.Writer.Header().Set(header.FrameOptions, header.DefaultFrameOptions)
	}
}
