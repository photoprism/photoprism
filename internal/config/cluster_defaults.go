package config

import (
	"net"
	"os"
	"regexp"
	"strings"
)

// getHostname is a var to allow tests to stub os.Hostname.
var getHostname = os.Hostname

// NonUniqueHostnames lists hostnames that must never be used as node name or to derive a cluster domain.
// It is a package variable on purpose so operators/tests can adjust in the future without spec changes.
var NonUniqueHostnames = map[string]struct{}{
	"localhost":             {},
	"localhost.localdomain": {},
	"localdomain":           {},
}

// ReservedDomains lists special/reserved domains that must not be used as cluster domains.
var ReservedDomains = map[string]struct{}{
	"example.com": {},
	"example.net": {},
	"example.org": {},
	"invalid":     {},
	"test":        {},
}

var dnsLabelRe = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$`)

// isDNSLabel returns true if s is a valid DNS label per our rules: lowercase, [a-z0-9-], 1–32 chars, starts/ends alnum.
func isDNSLabel(s string) bool {
	if s == "" || len(s) > 32 {
		return false
	}
	return dnsLabelRe.MatchString(s)
}

// isLocalSuffix returns true for .local mDNS or similar local-only suffixes we want to ignore.
func isLocalSuffix(suffix string) bool {
	return suffix == "local" || strings.HasSuffix(suffix, ".local")
}

// isDNSDomain validates a DNS domain (FQDN or single label not allowed here). It must have at least one dot.
// Each label must match isDNSLabel (except overall length and hyphen rules already covered by regex logic).
func isDNSDomain(d string) bool {
	d = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(d)), ".")
	if d == "" || strings.Count(d, ".") < 1 || len(d) > 253 {
		return false
	}
	if _, bad := ReservedDomains[d]; bad {
		return false
	}
	if isLocalSuffix(d) {
		return false
	}
	parts := strings.Split(d, ".")
	for _, p := range parts {
		if !isDNSLabel(p) {
			return false
		}
	}
	return true
}

// deriveSystemDomain tries to determine a usable cluster domain from system configuration.
// It uses the system hostname and returns the domain (everything after the first dot) when valid and not reserved.
func deriveSystemDomain() string {
	hn, _ := getHostname()
	hn = strings.ToLower(strings.TrimSpace(hn))
	if hn == "" {
		return ""
	}
	if _, bad := NonUniqueHostnames[hn]; bad {
		return ""
	}
	// If hostname contains a dot, take the domain part.
	if i := strings.IndexByte(hn, '.'); i > 0 && i < len(hn)-1 {
		dom := hn[i+1:]
		if isDNSDomain(dom) {
			return dom
		}
	}
	// Try reverse lookup to get FQDN domain, then validate.
	if addrs, err := net.LookupAddr(hn); err == nil {
		for _, fqdn := range addrs {
			fqdn = strings.TrimSuffix(strings.ToLower(fqdn), ".")
			if fqdn == "" || fqdn == hn {
				continue
			}
			if i := strings.IndexByte(fqdn, '.'); i > 0 && i < len(fqdn)-1 {
				dom := fqdn[i+1:]
				if isDNSDomain(dom) {
					return dom
				}
			}
		}
	}
	return ""
}
