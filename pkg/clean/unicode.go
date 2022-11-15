package clean

import "strings"

// Unicode returns a string a valid unicode.
func Unicode(s string) string {
	if s == "" {
		return ""
	}

	var b strings.Builder

	for _, c := range s {
		if c == '\uFFFD' {
			continue
		}
		b.WriteRune(c)
	}

	return b.String()
}
