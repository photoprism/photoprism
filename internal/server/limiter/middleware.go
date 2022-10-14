package limiter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Middleware registers the IP rate limiter middleware.
func Middleware(ip *Limit) gin.HandlerFunc {
	return func(c *gin.Context) {
		if l := ip.IP(c.ClientIP()); !l.Allow() {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
	}
}
