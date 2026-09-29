package proxy

import (
	"net"
	"net/http"
	"strings"

	"github.com/photoprism/photoprism/pkg/http/scheme"
)

// ParseTrustedProxies parses trusted proxy IP addresses and CIDR ranges as gin's
// SetTrustedProxies does. It returns nil if any value is invalid, since the server
// then trusts no proxy.
func ParseTrustedProxies(values []string) []*net.IPNet {
	result := make([]*net.IPNet, 0, len(values))

	for _, value := range values {
		if !strings.Contains(value, "/") {
			ip := net.ParseIP(value)

			if ip4 := ip.To4(); ip4 != nil {
				value += "/32"
			} else if ip != nil {
				value += "/128"
			} else {
				return nil
			}
		}

		_, network, err := net.ParseCIDR(value)

		if err != nil {
			return nil
		}

		result = append(result, network)
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// TrustedIP reports whether ip is in one of the trusted ranges.
func TrustedIP(ip string, trusted []*net.IPNet) bool {
	addr := net.ParseIP(ip)

	if addr == nil {
		return false
	}

	for _, network := range trusted {
		if network != nil && network.Contains(addr) {
			return true
		}
	}

	return false
}

// TrustedPeer reports whether the host of remoteAddr is in one of the trusted ranges.
func TrustedPeer(remoteAddr string, trusted []*net.IPNet) bool {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))

	if err != nil {
		return false
	}

	return TrustedIP(host, trusted)
}

// TrustedRequest reports whether the request was received from a trusted proxy,
// which includes any peer on a Unix socket.
func TrustedRequest(req *http.Request, trusted []*net.IPNet) bool {
	if req == nil {
		return false
	}

	if addr, ok := req.Context().Value(http.LocalAddrContextKey).(net.Addr); ok && addr != nil && addr.Network() == "unix" {
		return true
	}

	return TrustedPeer(req.RemoteAddr, trusted)
}

// ForwardedProto returns "https" if the request used TLS or, when sent by a trusted
// peer, carries one of the HTTPS protocol headers, and "http" otherwise.
func ForwardedProto(req *http.Request, trusted bool, httpsHeaders map[string]string) string {
	switch {
	case req == nil:
		return ""
	case req.TLS != nil:
		return scheme.Https
	case !trusted:
		return scheme.Http
	}

	for name, value := range httpsHeaders {
		v := req.Header.Get(name)

		if comma := strings.IndexByte(v, ','); comma >= 0 {
			v = v[:comma]
		}

		if v = strings.TrimSpace(v); v == "" {
			continue
		} else if strings.EqualFold(v, scheme.Https) || value != "" && strings.EqualFold(v, value) {
			return scheme.Https
		}
	}

	return scheme.Http
}
