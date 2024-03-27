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
