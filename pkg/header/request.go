package header

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/pkg/clean"
)

const (
	UnknownIP = "0.0.0.0"
	LocalIP   = "127.0.0.1"
)

// ClientIP returns the client IP address from the request context or a placeholder if it is unknown.
func ClientIP(c *gin.Context) (ip string) {
	if c == nil {
		// Should never happen.
		return UnknownIP
	} else if c.Request == nil {
		return UnknownIP
	} else if ip = c.ClientIP(); ip != "" {
		return clean.IP(ip, UnknownIP)
	} else if ip = c.RemoteIP(); ip != "" {
		return clean.IP(ip, UnknownIP)
	}

	// Tests may not specify an IP address.
	return UnknownIP
}

// UserAgent returns the user agent from the request context or an empty string if it is unknown.
func UserAgent(c *gin.Context) string {
	if c == nil {
		// Should never happen.
		return ""
	} else if c.Request == nil {
		return ""
	}

	return c.Request.UserAgent()
}
