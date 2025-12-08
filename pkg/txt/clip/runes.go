package clip

import "strings"

// Runes limits a string to the given number of runes and removes all leading and trailing spaces.
// Fast paths avoid []rune allocation when the input is ASCII and/or already within size.
func Runes(s string, size int) string {
	s = strings.TrimSpace(s)

	if s == "" || size <= 0 {
		return ""
	}

	// If length in bytes <= size, the string cannot exceed size runes.
	if len(s) <= size {
		return s
	}

	// ASCII fast path: byte length equals rune count → safe to slice by bytes.
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			ascii = false
			break
		}
	}
	if ascii {
		return strings.TrimSpace(s[:size])
	}

	// Fallback: count runes and slice exactly at rune boundary.
	runes := []rune(s)
	if len(runes) > size {
		s = string(runes[:size])
	}
	return strings.TrimSpace(s)
}
