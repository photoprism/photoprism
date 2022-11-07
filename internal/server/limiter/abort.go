package limiter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Abort cancels the request with error 429 (too many requests).
func Abort(c *gin.Context) {
	c.AbortWithStatus(http.StatusTooManyRequests)
}

// AbortJSON cancels the request with error 429 (too many requests).
func AbortJSON(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded", "code": http.StatusTooManyRequests})
}
