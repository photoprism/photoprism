package clean

import (
	"strings"
	"unicode"
)

// DNSLabel normalizes a string to a DNS label per our rules:
// - lowercase
// - allowed chars: [a-z0-9-]
// - other runes (including separators like space, '_', '.', '/', ':') map to '-'
// - collapses multiple '-' and trims leading/trailing '-'
// - maximum length 32 characters; trimming preserves start/end as alnum when possible
// Returns an empty string if no valid characters remain after normalization.
func DNSLabel(s string) string {
	if s == "" {
		return ""
	}

	const maxLen = 32
	s = strings.ToLower(s)

	var b strings.Builder
	b.Grow(len(s))
	prevDash := false

	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			if r == '-' {
				if prevDash {
					continue
				}
				prevDash = true
			} else {
				prevDash = false
			}
			b.WriteRune(r)
		default:
			// Treat any separator/invalid as a single '-'.
			// Includes space, underscore, dot, slash, colon, and others.
			if unicode.IsSpace(r) || r != 0 {
				if !prevDash {
					b.WriteByte('-')
					prevDash = true
				}
			}
		}
	}

	out := strings.Trim(b.String(), "-")
	if out == "" {
		return ""
	}

	if len(out) > maxLen {
		out = out[:maxLen]
		out = strings.Trim(out, "-")
		if out == "" {
			return ""
		}
	}
	return out
}
