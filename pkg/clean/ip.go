package clean

import (
	"net"
	"regexp"
)

// IpRegExp matches characters allowed in IPv4 or IPv6 network addresses.
var IpRegExp = regexp.MustCompile(`[^a-zA-Z0-9:.]`)

// IP returns the sanitized and normalized network address if it is valid, or the default otherwise.
func IP(s, defaultIp string) string {
	// Return default if invalid.
	if s == "" || len(s) > LengthLimit || s == defaultIp {
		return defaultIp
	}

	// Remove invalid characters, including whitespace.
	if s = IpRegExp.ReplaceAllString(s, ""); s == "" {
		return defaultIp
	}

	// Limit string length to 39 characters.
	if len(s) > LengthIPv6 {
		s = s[:LengthIPv6]
	}

	// Parse IP address and return it as string.
	if ip := net.ParseIP(s); ip == nil {
		return defaultIp
	} else {
		return ip.String()
	}
}
