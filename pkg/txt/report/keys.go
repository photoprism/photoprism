package report

import (
	"regexp"
	"strings"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9_]+`)
var underscores = regexp.MustCompile(`_+`)

// CanonKey converts a column title into a stable snake_case key suitable
// for JSON output. It lowercases, replaces spaces/hyphens/slashes with '_',
// removes other punctuation, collapses repeats, and trims edges.
func CanonKey(s string) string {
	k := strings.ToLower(s)
	k = strings.NewReplacer(" ", "_", "-", "_", "/", "_").Replace(k)
	k = nonAlnum.ReplaceAllString(k, "_")
	k = underscores.ReplaceAllString(k, "_")
	k = strings.Trim(k, "_")
	if k == "" {
		return "col"
	}
	return k
}
