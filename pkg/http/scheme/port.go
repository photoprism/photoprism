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

// NormalizeBaseURL returns s as a base URL with a trailing slash, stripping
// the scheme's default port, query strings, and fragments. Userinfo is
// preserved. Returns s with a trailing slash appended if parsing fails.
func NormalizeBaseURL(s string) string {
	u, err := url.Parse(s)

	if err != nil {
		return strings.TrimRight(s, "/") + "/"
	}

	StripDefaultPort(u)
	u.RawQuery = ""
	u.ForceQuery = false
	u.Fragment = ""
	u.RawFragment = ""
	u.Path = strings.TrimRight(u.Path, "/") + "/"

	return u.String()
}
