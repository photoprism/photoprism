package dns

import (
	"net"
	"os"
	"strings"
)

// GetHostname is a var to allow tests to stub os.Hostname.
var GetHostname = os.Hostname

// IsDomain validates a DNS domain (FQDN or single label not allowed here). It must have at least one dot.
// Each label must match IsLabel (except overall length and hyphen rules already covered by regex logic).
func IsDomain(d string) bool {
	d = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(d)), ".")
	if d == "" || strings.Count(d, ".") < 1 || len(d) > 253 {
		return false
	}
	if _, bad := ReservedDomains[d]; bad {
		return false
	}
	if IsLocalSuffix(d) {
		return false
	}
	for p := range strings.SplitSeq(d, ".") {
		if !IsLabel(p) {
			return false
		}
	}
	return true
}

// GetSystemDomain tries to determine a usable cluster domain from system configuration.
// It uses the system hostname and returns the domain (everything after the first dot) when valid and not reserved.
func GetSystemDomain() string {
	hn, _ := GetHostname()
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
		if IsDomain(dom) {
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
				if IsDomain(dom) {
					return dom
				}
			}
		}
	}

	return ""
}
