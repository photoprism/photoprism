package limiter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/pkg/http/header"
)

// Middleware registers the IP rate limiter middleware.
func Middleware(limiter *Limit) gin.HandlerFunc {
	return func(c *gin.Context) {
		if l := limiter.IP(header.ClientIP(c)); !l.Allow() {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
	}
}
