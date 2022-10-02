package api

import (
	"github.com/gin-gonic/gin"
)

const UnknownIP = "0.0.0.0"

// ClientIP returns the client IP address from the request context or a placeholder if it is unknown.
func ClientIP(c *gin.Context) (ip string) {
	if c == nil {
		// Should never happen.
		return UnknownIP
	} else if ip = c.ClientIP(); ip == "" {
		// Unit tests often do not set a client IP.
		return UnknownIP
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
