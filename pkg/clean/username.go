package clean

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Username returns the normalized username (lowercase, whitespace trimmed).
func Username(s string) string {
	s = strings.TrimSpace(s)

	if s == "" || reject(s, txt.ClipUsername) {
		return ""
	}

	return strings.ToLower(s)
}
