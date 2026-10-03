package header

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

// Optional HTTP request header names.
const (
	Cookie    = "Cookie"
	Referer   = "Referer"
	Browser   = "Sec-Ch-Ua"
	Platform  = "Sec-Ch-Ua-Platform"
	FetchMode = "Sec-Fetch-Mode"
	UserAgent = "User-Agent"
)

// Standard IP addresses and placeholders.
const (
	UnknownIP = "0.0.0.0"
	LocalIP   = "127.0.0.1"
)

// forwardedForPlatform reports whether X-Forwarded-For is the trusted platform header.
var forwardedForPlatform atomic.Bool

// SetTrustedPlatform sets the trusted platform header and returns the name gin should resolve,
// which is empty for X-Forwarded-For, as ClientIP reads that header from all lines.
func SetTrustedPlatform(name string) string {
	name = strings.TrimSpace(name)
	isForwardedFor := strings.EqualFold(name, XForwardedFor)
	forwardedForPlatform.Store(isForwardedFor)

	if isForwardedFor {
		return ""
	}

	return name
}

// ClientIP returns the client IP address from the request context or a placeholder if it is unknown.
// With X-Forwarded-For as trusted platform header, it is the last entry, or the peer if that is invalid.
func ClientIP(c *gin.Context) (ip string) {
	if c == nil {
		// Should never happen.
		return UnknownIP
	} else if c.Request == nil {
		return UnknownIP
	} else if last, found := forwardedForIP(c); found {
		return last
	} else if ip = c.ClientIP(); ip != "" {
		return IP(ip, UnknownIP)
	} else if ip = c.RemoteIP(); ip != "" {
		return IP(ip, UnknownIP)
	}

	// Tests may not specify an IP address.
	return UnknownIP
}

// forwardedForIP returns the last X-Forwarded-For entry, or the peer if it is invalid,
// and false if X-Forwarded-For is not the trusted platform header or is missing.
func forwardedForIP(c *gin.Context) (string, bool) {
	if !forwardedForPlatform.Load() {
		return "", false
	}

	last, found := LastValue(c.Request.Header, XForwardedFor)

	if !found {
		return "", false
	} else if ip, err := ParseIP(last); err == nil {
		return ip, true
	}

	return IP(c.RemoteIP(), UnknownIP), true
}

// LastValue returns the last comma-separated entry across all lines of the named header,
// and false if the header is missing or has only blank lines.
func LastValue(h http.Header, name string) (string, bool) {
	var value string

	for _, line := range h.Values(name) {
		if line = strings.TrimSpace(line); line != "" {
			value = line
		}
	}

	if value == "" {
		return "", false
	} else if i := strings.LastIndexByte(value, ','); i >= 0 {
		value = value[i+1:]
	}

	return strings.TrimSpace(value), true
}

// ClientUserAgent returns the client user agent string
// from the request context, or an empty string if unknown.
func ClientUserAgent(c *gin.Context) string {
	if c == nil {
		// Should never happen.
		return ""
	} else if c.Request == nil {
		return ""
	}

	return c.Request.UserAgent()
}
