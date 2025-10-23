package dns

import (
	"regexp"
)

var labelRegex = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$`)

// IsLabel returns true if s is a valid DNS label per our rules: lowercase, [a-z0-9-], 1–32 chars, starts/ends alnum.
func IsLabel(s string) bool {
	if s == "" || len(s) > 32 {
		return false
	}

	return labelRegex.MatchString(s)
}
