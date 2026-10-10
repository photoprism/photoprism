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
	addr, err := ParseAddr(s)

	if err != nil {
		return "", err
	}

	return addr.String(), nil
}

// ParseAddr parses a single network address as ParseIP does and returns it as netip.Addr.
func ParseAddr(s string) (netip.Addr, error) {
	if len(s) > MaxIPLength {
		return netip.Addr{}, ErrInvalidIP
	} else if s = strings.TrimSpace(s); s == "" {
		return netip.Addr{}, ErrInvalidIP
	}

	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = s[1 : len(s)-1]
	} else if strings.HasPrefix(s, "[") || strings.Count(s, ":") == 1 {
		host, port, err := net.SplitHostPort(s)

		if err != nil || !isPort(port) {
			return netip.Addr{}, ErrInvalidIP
		}

		s = host
	}

	addr, err := netip.ParseAddr(s)

	if err != nil {
		return netip.Addr{}, ErrInvalidIP
	}

	return addr.WithZone("").Unmap(), nil
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

// IPv6NetworkBits is the prefix length of the network a global IPv6 client is counted in.
const IPv6NetworkBits = 64

var (
	// ipv6Global is the global unicast range, whose clients are counted by network.
	ipv6Global = netip.MustParsePrefix("2000::/3")
	// ipv6Teredo addresses embed the client's IPv4 address, inverted, in their last 32 bits.
	ipv6Teredo = netip.MustParsePrefix("2001::/32")
	// ipv6Nat64 addresses embed the IPv4 address of a translated client in their last 32 bits.
	ipv6Nat64 = netip.MustParsePrefix("64:ff9b::/96")
	// ipv66to4 addresses embed the IPv4 address of the site in bits 16 to 47.
	ipv66to4 = netip.MustParsePrefix("2002::/16")
	// ipv4Shared is the shared address space for carrier-grade NAT.
	ipv4Shared = netip.MustParsePrefix("100.64.0.0/10")
)

// ClientNetwork returns the network a client address is counted in: the address itself for IPv4,
// the embedded global IPv4 address for NAT64, the embedded IPv4 address with a "teredo:" or "6to4:"
// prefix for those tunnels, the /64 prefix for other global IPv6 addresses, and the address itself
// otherwise, or an empty string if the address is invalid.
func ClientNetwork(ip string) string {
	addr, err := ParseAddr(ip)

	if err != nil {
		return ""
	}

	switch {
	case addr.Is4():
		return addr.String()
	case ipv6Nat64.Contains(addr):
		b := addr.As16()

		if v4 := netip.AddrFrom4([4]byte(b[12:])); isGlobalIPv4(v4) {
			return v4.String()
		}

		return addr.String()
	case ipv6Teredo.Contains(addr):
		b := addr.As16()
		return "teredo:" + netip.AddrFrom4([4]byte{^b[12], ^b[13], ^b[14], ^b[15]}).String()
	case ipv66to4.Contains(addr):
		b := addr.As16()
		return "6to4:" + netip.AddrFrom4([4]byte(b[2:6])).String()
	case ipv6Global.Contains(addr):
		return netip.PrefixFrom(addr, IPv6NetworkBits).Masked().String()
	default:
		return addr.String()
	}
}

// isGlobalIPv4 reports whether addr is a globally routed IPv4 address.
func isGlobalIPv4(addr netip.Addr) bool {
	return addr.Is4() && addr.IsGlobalUnicast() && !addr.IsPrivate() && !ipv4Shared.Contains(addr)
}
