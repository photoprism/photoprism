package cluster

import (
	"errors"
	"net/netip"
	"strings"

	"github.com/photoprism/photoprism/pkg/http/header"
)

// ErrInvalidCIDR is returned for a cluster CIDR list with an empty or invalid entry.
var ErrInvalidCIDR = errors.New("invalid cluster cidr")

// ParseCIDRs parses a comma-separated list of CIDR ranges, such as the cluster-cidr option,
// and returns ErrInvalidCIDR if the list is empty or any entry is invalid.
func ParseCIDRs(value string) ([]netip.Prefix, error) {
	var result []netip.Prefix

	for _, entry := range strings.Split(value, ",") {
		prefix, err := netip.ParsePrefix(trimPrefixBits(strings.TrimSpace(entry)))

		if err != nil {
			return nil, ErrInvalidCIDR
		}

		// Write an IPv4-mapped range as the IPv4 range it covers, as client addresses are unmapped.
		if addr := prefix.Addr(); addr.Is4In6() && prefix.Bits() >= 96 {
			prefix = netip.PrefixFrom(addr.Unmap(), prefix.Bits()-96)
		}

		result = append(result, prefix.Masked())
	}

	return result, nil
}

// trimPrefixBits removes leading zeros from the prefix length of a CIDR range, e.g. "/08".
func trimPrefixBits(s string) string {
	i := strings.LastIndexByte(s, '/')

	if i < 0 || i == len(s)-1 {
		return s
	}

	if bits := strings.TrimLeft(s[i+1:], "0"); bits == "" {
		return s[:i+1] + "0"
	} else {
		return s[:i+1] + bits
	}
}

// CIDRsContain reports whether the client address is inside one of the comma-separated CIDR ranges,
// and false if the list is invalid or the address cannot be parsed or is unspecified.
func CIDRsContain(value, clientIP string) bool {
	prefixes, err := ParseCIDRs(value)

	if err != nil {
		return false
	}

	addr, err := header.ParseAddr(clientIP)

	if err != nil || addr.IsUnspecified() {
		return false
	}

	for _, prefix := range prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}

	return false
}
