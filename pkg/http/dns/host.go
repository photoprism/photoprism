package dns

import (
	"net"
	"strconv"
	"strings"
)

// TrimBrackets returns host without the square brackets of a bracketed IPv6 literal.
func TrimBrackets(host string) string {
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return host[1 : len(host)-1]
	}

	return host
}

// JoinHostPort joins a host name or address, bracketed or not, and a port into host:port,
// with an IPv6 literal in brackets. It does not validate host.
func JoinHostPort(host string, port int) string {
	return net.JoinHostPort(TrimBrackets(host), strconv.Itoa(port))
}

// BracketHost returns host with an IPv6 literal in brackets, as it appears in a URL or Host header.
// It does not validate host.
func BracketHost(host string) string {
	if host = TrimBrackets(host); strings.Contains(host, ":") {
		return "[" + host + "]"
	}

	return host
}
