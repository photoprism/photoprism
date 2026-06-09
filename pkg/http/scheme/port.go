package scheme

import (
	"net/url"
	"strings"
)

// DefaultPort returns the IANA default port for a known URL scheme, or "" if unknown.
func DefaultPort(s string) string {
	switch s {
	case Http:
		return "80"
	case Https, Websocket:
		return "443"
	}

	return ""
}

// StripDefaultPort removes the explicit port from u.Host when it equals the
// default for u.Scheme. IPv6 brackets are preserved.
func StripDefaultPort(u *url.URL) {
	if u == nil {
		return
	}

	port := u.Port()

	if port == "" || port != DefaultPort(u.Scheme) {
		return
	}

	u.Host = strings.TrimSuffix(u.Host, ":"+port)
}
