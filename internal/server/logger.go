package server

import (
	"time"

	"github.com/gin-gonic/gin"
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

		// clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		if statusCode >= 400 {
			log.Errorf("%s %s (%3d) [%v]",
				method,
				path,
				statusCode,
				latency,
			)
		} else {
			log.Debugf("%s %s (%3d) [%v]",
				method,
				path,
				statusCode,
				latency,
			)
		}
	}
}
