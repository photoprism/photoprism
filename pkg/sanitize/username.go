package sanitize

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Username returns the normalized username (lowercase, whitespace trimmed).
func Username(s string) string {
	return strings.ToLower(txt.Clip(s, txt.ClipUsername))
}
