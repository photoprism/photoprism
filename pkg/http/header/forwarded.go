package header

import (
	"net/http"
	"strings"
)

// ForwardedHeaders lists inbound request headers that carry a client address,
// host, protocol, or path prefix and are not passed on to upstream servers.
var ForwardedHeaders = []string{
	Forwarded,
	XForwardedFor,
	XForwardedHost,
	XForwardedProto,
	XForwardedPort,
	XForwardedPrefix,
	XForwardedServer,
	XForwardedSsl,
	XOriginalForwardedFor,
	XUrlScheme,
	FrontEndHttps,
	XRealIP,
	XClientIP,
	CFConnectingIP,
	FlyClientIP,
	XAppengineRemoteAddr,
	TrueClientIP,
	XClusterClientIP,
	FastlyClientIP,
}

// DeleteForwarded removes ForwardedHeaders and the extra header names from h.
func DeleteForwarded(h http.Header, extra ...string) {
	if h == nil {
		return
	}

	for _, name := range ForwardedHeaders {
		h.Del(name)
	}

	for _, name := range extra {
		if name = strings.TrimSpace(name); name != "" {
			h.Del(name)
		}
	}
}
