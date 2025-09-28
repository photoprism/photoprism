package clean

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/list"
)

// Scope sanitizes a string that contains auth scope identifiers.
// Callers should use acl.ScopeAttrPermits / acl.ScopePermits for authorization checks.
func Scope(s string) string {
	if s == "" {
		return ""
	}

	return list.ParseAttr(strings.ToLower(s)).String()
}

// Scopes sanitizes auth scope identifiers and returns them as strings.
// Callers should use acl.ScopeAttrPermits / acl.ScopePermits for authorization checks.
func Scopes(s string) []string {
	if s == "" {
		return []string{}
	}

	return list.ParseAttr(strings.ToLower(s)).Strings()
}
