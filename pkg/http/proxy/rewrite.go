package proxy

import (
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/pkg/http/dns"
)

// portalRootPathPrefixes lists URL paths that are served by the Portal root,
// not by any proxied instance. Locations under these prefixes are deliberate
// instance-to-Portal redirects (for example the Pro OIDC RP redirecting to
// the Portal's authorize endpoint), and the per-instance path prefix must
// not be added to them. The OIDC OP endpoints live under "/api/v1/oauth/"
// specifically (not all of "/api/v1/", which is an instance surface too).
var portalRootPathPrefixes = []string{
	"/api/v1/oauth/",
	"/.well-known/",
	"/portal/",
}

// isPortalRootPath reports whether a URL path targets a Portal-root surface.
func isPortalRootPath(p string) bool {
	if p == "" {
		return false
	}
	scoped := "/" + strings.TrimLeft(p, "/")
	for _, prefix := range portalRootPathPrefixes {
		if scoped == strings.TrimSuffix(prefix, "/") || strings.HasPrefix(scoped, prefix) {
			return true
		}
	}
	return false
}

// RewriteLocation prefixes redirect targets to keep clients within a proxy path scope.
func RewriteLocation(location, pathPrefix, proxyHost string) string {
	if location == "" || pathPrefix == "" {
		return location
	}

	if strings.HasPrefix(location, "/") {
		if HasPathPrefix(location, pathPrefix) || isPortalRootPath(location) {
			return location
		}
		return JoinPathPrefix(pathPrefix, location)
	}

	u, err := url.Parse(location)

	if err != nil || u.Host == "" || proxyHost == "" {
		return location
	}

	if !HostMatch(u.Host, proxyHost) {
		return location
	}

	if HasPathPrefix(u.Path, pathPrefix) || isPortalRootPath(u.Path) {
		return location
	}

	u.Path = JoinPathPrefix(pathPrefix, u.Path)
	u.RawPath = ""

	return u.String()
}

// HasPathPrefix checks whether a URL path is already scoped to a proxy prefix.
func HasPathPrefix(pathValue, pathPrefix string) bool {
	scopedPrefix := "/" + strings.Trim(pathPrefix, "/")
	scopedPath := "/" + strings.TrimLeft(pathValue, "/")

	if scopedPrefix == "/" {
		return true
	}

	return scopedPath == scopedPrefix || strings.HasPrefix(scopedPath, scopedPrefix+"/")
}

// JoinPathPrefix joins a proxy prefix and path without duplicating slashes.
func JoinPathPrefix(pathPrefix, pathValue string) string {
	scopedPrefix := "/" + strings.Trim(pathPrefix, "/")
	scopedPath := strings.TrimLeft(pathValue, "/")

	if scopedPath == "" {
		return scopedPrefix + "/"
	}

	return scopedPrefix + "/" + scopedPath
}

// RewriteSetCookiePath enforces cookie Path scoping for proxy-routed requests.
func RewriteSetCookiePath(value, pathPrefix string) string {
	if value == "" || pathPrefix == "" {
		return value
	}

	parts := strings.Split(value, ";")

	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if len(trimmed) < 5 {
			continue
		}
		if !strings.EqualFold(trimmed[:5], "path=") {
			continue
		}

		pathValue := strings.TrimSpace(trimmed[5:])
		if pathValue == "" || pathValue == "/" {
			parts[i] = " Path=" + pathPrefix
			return strings.Join(parts, ";")
		}

		return value
	}

	return value + "; Path=" + pathPrefix
}

// HostMatch compares hosts while tolerating optional ports, and reports false if either is not a valid
// host[:port], where only an IPv6 address may be in brackets and a port must be numeric.
func HostMatch(a, b string) bool {
	aHost, aOk := hostKey(a)
	bHost, bOk := hostKey(b)

	return aOk && bOk && aHost == bHost
}

// hostKey returns the host of a host[:port] value in comparable form: an IP address in canonical form
// or a lowercase ASCII name. It reports false for an empty, malformed, or non-ASCII value.
func hostKey(s string) (string, bool) {
	host := s

	if h, port, err := net.SplitHostPort(s); err == nil {
		if n, convErr := strconv.ParseUint(port, 10, 16); convErr != nil || n < 1 || port[0] == '0' {
			return "", false
		}

		host = h
	} else if bracketed := dns.TrimBrackets(s); bracketed != s {
		host = bracketed
	}

	if strings.HasPrefix(s, "[") || strings.Contains(host, ":") {
		addr, err := netip.ParseAddr(host)

		if err != nil || !addr.Is6() || addr.Is4In6() {
			return "", false
		}

		return addr.WithZone("").String(), true
	} else if addr, err := netip.ParseAddr(host); err == nil {
		return addr.String(), true
	} else if host == "" || strings.ContainsAny(host, "[]@/ ") || !dns.IsASCII(host) {
		return "", false
	}

	return strings.ToLower(host), true
}

// RewriteDestinationHost rewrites absolute WebDAV Destination headers from a proxy host to the upstream host.
func RewriteDestinationHost(req *http.Request, proxyHost string, upstream *url.URL) {
	if req == nil || upstream == nil {
		return
	}

	raw := strings.TrimSpace(req.Header.Get("Destination"))

	if raw == "" {
		return
	}

	u, err := url.Parse(raw)

	if err != nil || u.Host == "" {
		return
	}

	if proxyHost == "" || !HostMatch(u.Host, proxyHost) {
		return
	}

	u.Scheme = upstream.Scheme
	u.Host = upstream.Host
	req.Header.Set("Destination", u.String())
}
