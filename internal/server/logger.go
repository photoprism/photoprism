package server

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Logger instances a Logger middleware for Gin.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		// clientIp := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// Use debug level to keep production logs clean.
		log.Debugf("server: %s %s (%3d) [%v]",
			method,
			clean.Log(path),
			statusCode,
			latency,
		)
	}
}
