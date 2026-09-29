package dns

import (
	"net"
	"strings"

	"golang.org/x/net/idna"
)

// MaxNameLength is the maximum length of a DNS name in ASCII form, without a trailing dot.
const MaxNameLength = 253

// nonPublicSuffixes lists special-use top-level names that are never delegated in public DNS.
var nonPublicSuffixes = []string{"example", "home.arpa", "internal", "invalid", "local", "localhost", "test"}

// IsPublicName reports whether name is a DNS name a public certificate authority can issue a certificate for.
// Names are compared in ASCII form, so internationalized names and labels up to 63 characters are accepted.
func IsPublicName(name string) bool {
	name = strings.TrimSuffix(strings.TrimSpace(name), ".")

	if name == "" || net.ParseIP(name) != nil {
		return false
	}

	ascii, err := idna.Lookup.ToASCII(name)

	if err != nil || len(ascii) > MaxNameLength || !strings.Contains(ascii, ".") {
		return false
	} else if _, reserved := ReservedDomains[ascii]; reserved {
		return false
	}

	for _, suffix := range nonPublicSuffixes {
		if strings.HasSuffix(ascii, "."+suffix) {
			return false
		}
	}

	labels := strings.Split(ascii, ".")

	for _, label := range labels {
		if !isPublicLabel(label) {
			return false
		}
	}

	// A numeric last label is read as an IPv4 address by URL parsers, e.g. 127.1.
	return strings.Trim(labels[len(labels)-1], "0123456789") != ""
}

// isPublicLabel reports whether s is a DNS label of 1 to 63 lowercase letters, digits, or inner hyphens.
func isPublicLabel(s string) bool {
	if s == "" || len(s) > 63 || s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}

	for i := 0; i < len(s); i++ {
		if c := s[i]; (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}

	return true
}
