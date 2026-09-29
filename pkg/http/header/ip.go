package header

import (
	"errors"
	"net"
	"net/netip"
	"strings"
)

// MaxIPLength is the maximum length of a network address string accepted by ParseIP.
const MaxIPLength = 64

// ErrInvalidIP is returned by ParseIP for input that is not a single network address.
var ErrInvalidIP = errors.New("invalid ip address")

// IsIP returns true if the string matches a valid IP address.
func IsIP(s string) bool {
	return IP(s, "") != ""
}

// IP returns the normalized network address if it is valid, or the default otherwise.
func IP(s, defaultIp string) string {
	if s == "" || s == defaultIp {
		return defaultIp
	} else if ip, err := ParseIP(s); err != nil {
		return defaultIp
	} else {
		return ip
	}
}

// ParseIP parses a single network address, optionally bracketed or with a port, and
// returns it in canonical form without zone, with IPv4-mapped addresses as IPv4.
func ParseIP(s string) (string, error) {
	if len(s) > MaxIPLength {
		return "", ErrInvalidIP
	} else if s = strings.TrimSpace(s); s == "" {
		return "", ErrInvalidIP
	}

	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = s[1 : len(s)-1]
	} else if strings.HasPrefix(s, "[") || strings.Count(s, ":") == 1 {
		host, port, err := net.SplitHostPort(s)

		if err != nil || !isPort(port) {
			return "", ErrInvalidIP
		}

		s = host
	}

	addr, err := netip.ParseAddr(s)

	if err != nil {
		return "", ErrInvalidIP
	}

	return addr.WithZone("").Unmap().String(), nil
}

// isPort reports whether s is a decimal port number.
func isPort(s string) bool {
	if s == "" || len(s) > 5 {
		return false
	}

	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}
