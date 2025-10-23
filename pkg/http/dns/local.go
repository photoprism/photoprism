package dns

import (
	"net"
	"strings"
)

// NonUniqueHostnames lists hostnames that must never be used as node name or to
// derive a cluster domain. It is mutable on purpose so tests or operators can
// extend the set without changing the package API.
var NonUniqueHostnames = map[string]struct{}{
	"localhost":             {},
	"localhost.localdomain": {},
	"localdomain":           {},
}

// IsLocalSuffix reports whether the provided suffix is considered local-only
// (for example mDNS domains ending in .local) and therefore unsuitable when
// deriving public cluster domains.
func IsLocalSuffix(suffix string) bool {
	return suffix == "local" || strings.HasSuffix(suffix, ".local")
}

// IsLoopbackHost reports whether host refers to a loopback address that is safe
// to contact over plain HTTP during bootstrap. It accepts hostnames (e.g.
// "localhost") as well as IPv4/IPv6 addresses and normalises case/whitespace.
func IsLoopbackHost(host string) bool {
	h := strings.TrimSpace(strings.ToLower(host))
	if h == "" {
		return false
	}

	if h == "localhost" {
		return true
	}

	if ip := net.ParseIP(h); ip != nil {
		return ip.IsLoopback()
	}

	return false
}
