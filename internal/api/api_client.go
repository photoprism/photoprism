package api

import (
	"github.com/gin-gonic/gin"
)

// ClientIP returns the client IP address from the request context or a placeholder if it is unknown.
func ClientIP(c *gin.Context) (ip string) {
	if c == nil {
		// Should never happen.
		return "0.0.0.0"
	} else if ip = c.ClientIP(); ip == "" {
		// Unit tests generally do not set a client IP. According to RFC 5737, the 192.0.2.0/24 subnet
		// is intended for use in documentation and examples.
		return "192.0.2.42"
	}

	return ip
}

// UserAgent returns the user agent from the request context or an empty string if it is unknown.
func UserAgent(c *gin.Context) string {
	if c == nil {
		// Should never happen.
		return ""
	}

	return c.Request.UserAgent()
}
