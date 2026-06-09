package proxy

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/photoprism/photoprism/pkg/http/header"
)

// ForwardedProto determines the forwarded scheme for proxy headers.
func ForwardedProto(req *http.Request) string {
	if req == nil {
		return ""
	}

	if v := strings.TrimSpace(req.Header.Get(header.XForwardedProto)); v != "" {
		if comma := strings.IndexByte(v, ','); comma > 0 {
			return strings.TrimSpace(v[:comma])
		}
		return v
	}

	if req.TLS != nil {
		return "https"
	}

	return "http"
}

// portalRootPathPrefixes lists URL paths that are served by the Portal root,
// not by any proxied instance. Locations under these prefixes are deliberate
// instance-to-Portal redirects (for example the Pro OIDC RP redirecting to
// the Portal's authorize endpoint), and the per-instance path prefix must
// not be added to them.
var portalRootPathPrefixes = []string{
	"/oauth/",
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

// HostMatch compares hosts while tolerating optional ports.
func HostMatch(a, b string) bool {
	if strings.EqualFold(a, b) {
		return true
	}

	aHost := a

	if host, _, err := net.SplitHostPort(a); err == nil && host != "" {
		aHost = host
	}

	bHost := b

	if host, _, err := net.SplitHostPort(b); err == nil && host != "" {
		bHost = host
	}

	return strings.EqualFold(aHost, bHost)
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
