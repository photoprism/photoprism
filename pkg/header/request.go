package header

import (
	"github.com/gin-gonic/gin"
)

// Optional HTTP request header names.
const (
	Cookie    = "Cookie"
	Referer   = "Referer"
	Browser   = "Sec-Ch-Ua"
	Platform  = "Sec-Ch-Ua-Platform"
	FetchMode = "Sec-Fetch-Mode"
)

// Standard IP addresses and placeholders.
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
		return IP(ip, UnknownIP)
	} else if ip = c.RemoteIP(); ip != "" {
		return IP(ip, UnknownIP)
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
