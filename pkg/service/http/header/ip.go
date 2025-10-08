package header

import (
	"net"
	"regexp"
)

// IpRegExp matches characters allowed in IPv4 or IPv6 network addresses.
// Kept for backwards compatibility (other packages reference it), but IP() no longer uses it.
var IpRegExp = regexp.MustCompile(`[^a-zA-Z0-9:.]`)

const (
	IPv6Length = 39
)

// IsIP returns true if the string matches a valid IP address.
func IsIP(s string) bool {
	return IP(s, "") != ""
}

// IP returns the sanitized and normalized network address if it is valid, or the default otherwise.
func IP(s, defaultIp string) string {
	// Return default if invalid.
	if s == "" || len(s) > 64 || s == defaultIp {
		return defaultIp
	}

	// Filter invalid characters: allow only [A-Za-z0-9:.]
	fastOK := true
	for i := 0; i < len(s); i++ {
		b := s[i]
		if !((b >= '0' && b <= '9') || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == ':' || b == '.') {
			fastOK = false
			break
		}
	}
	if !fastOK {
		dst := make([]byte, 0, len(s))
		for i := 0; i < len(s); i++ {
			c := s[i]
			if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == ':' || c == '.' {
				dst = append(dst, c)
			}
		}
		s = string(dst)
		if s == "" {
			return defaultIp
		}
	}

	// Limit string length to 39 characters.
	if len(s) > IPv6Length {
		s = s[:IPv6Length]
	}

	// Parse IP address and return it as string.
	if ip := net.ParseIP(s); ip == nil {
		return defaultIp
	} else {
		return ip.String()
	}
}
