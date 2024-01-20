package limiter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Middleware registers the IP rate limiter middleware.
func Middleware(limiter *Limit) gin.HandlerFunc {
	return func(c *gin.Context) {
		if l := limiter.IP(c.ClientIP()); !l.Allow() {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
	}
}
