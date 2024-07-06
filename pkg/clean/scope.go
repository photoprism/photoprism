package clean

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/list"
)

// Scope sanitizes a string that contains authentication scope identifiers.
func Scope(s string) string {
	if s == "" {
		return ""
	}

	return list.ParseAttr(strings.ToLower(s)).String()
}

// Scopes sanitizes authentication scope identifiers and returns them as string slice.
func Scopes(s string) []string {
	if s == "" {
		return []string{}
	}

	return list.ParseAttr(strings.ToLower(s)).Strings()
}
