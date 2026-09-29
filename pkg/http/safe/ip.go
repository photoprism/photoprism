package safe

import (
	"context"
	"net"
	"net/netip"
	"syscall"
	"time"
)

// lookupIPAddr resolves a host name; tests may replace it.
var lookupIPAddr = net.DefaultResolver.LookupIPAddr

// dialControlFunc checks each connection of a download that disallows private networks, and
// disallowedPeer the connected peer address afterwards; tests may replace them.
var (
	dialControlFunc = dialControl
	disallowedPeer  = isPrivateOrDisallowedIP
)

var (
	// disallowedPrefixes lists special-purpose ranges that are refused besides the private, loopback,
	// link-local, multicast, and unspecified addresses.
	disallowedPrefixes = []netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"),      // This network.
		netip.MustParsePrefix("100.64.0.0/10"),  // Shared address space (CGNAT).
		netip.MustParsePrefix("240.0.0.0/4"),    // Reserved, including broadcast.
		netip.MustParsePrefix("::/96"),          // IPv4-compatible, including unspecified and loopback.
		netip.MustParsePrefix("100::/64"),       // Discard-only.
		netip.MustParsePrefix("2001::/32"),      // Teredo.
		netip.MustParsePrefix("64:ff9b:1::/48"), // Local-use IPv4/IPv6 translation.
		netip.MustParsePrefix("fec0::/10"),      // Site-local.
	}
	// nat64Prefix embeds an IPv4 address in its last 32 bits.
	nat64Prefix = netip.MustParsePrefix("64:ff9b::/96")
	// sixToFourPrefix embeds an IPv4 address in bits 16 to 47.
	sixToFourPrefix = netip.MustParsePrefix("2002::/16")
)

// isPrivateOrDisallowedIP reports whether ip is not a public unicast address, or is invalid.
func isPrivateOrDisallowedIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)

	if !ok {
		return true
	}

	return disallowedAddr(addr)
}

// disallowedAddr reports whether addr is not a public unicast address, including an IPv4 address
// embedded in a NAT64 or 6to4 address.
func disallowedAddr(addr netip.Addr) bool {
	addr = addr.WithZone("").Unmap()

	if !addr.IsValid() || addr.IsUnspecified() || addr.IsLoopback() || addr.IsMulticast() ||
		addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsInterfaceLocalMulticast() ||
		addr.IsPrivate() || !addr.IsGlobalUnicast() {
		return true
	}

	for _, p := range disallowedPrefixes {
		if p.Contains(addr) {
			return true
		}
	}

	switch b := addr.As16(); {
	case nat64Prefix.Contains(addr):
		return disallowedAddr(netip.AddrFrom4([4]byte(b[12:])))
	case sixToFourPrefix.Contains(addr):
		return disallowedAddr(netip.AddrFrom4([4]byte(b[2:6])))
	}

	return false
}

// checkHost returns ErrPrivateIP if host is, or resolves to, an address that is not allowed.
func checkHost(host string) error {
	if ip := net.ParseIP(host); ip != nil {
		if isPrivateOrDisallowedIP(ip) {
			return ErrPrivateIP
		}

		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addrs, err := lookupIPAddr(ctx, host)

	if err != nil {
		return err
	}

	for _, a := range addrs {
		if isPrivateOrDisallowedIP(a.IP) {
			return ErrPrivateIP
		}
	}

	return nil
}

// dialControl refuses a connection to an address that is not allowed before it is opened.
func dialControl(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)

	if err != nil {
		return ErrPrivateIP
	} else if ip := net.ParseIP(host); ip == nil || isPrivateOrDisallowedIP(ip) {
		return ErrPrivateIP
	}

	return nil
}
