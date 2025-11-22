package dns

// ReservedDomains lists special/reserved domains that must not be used as cluster domains.
var ReservedDomains = map[string]struct{}{
	"example.com": {},
	"example.net": {},
	"example.org": {},
	"invalid":     {},
	"test":        {},
}
