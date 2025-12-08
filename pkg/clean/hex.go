package clean

import (
	"strings"
)

// Hex removes invalid characters from a hex string and lowercases A-F.
func Hex(s string) string {
	if s == "" || reject(s, 1024) {
		return ""
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// Scan once; lower-case A-F on the fly; drop non-hex.
	// Allocate only if needed.
	var out []byte
	for i := 0; i < len(s); i++ {
		b := s[i]
		switch {
		case b >= '0' && b <= '9':
			if out != nil {
				out = append(out, b)
			}
		case b >= 'a' && b <= 'f':
			if out != nil {
				out = append(out, b)
			}
		case b >= 'A' && b <= 'F':
			if out == nil {
				out = make([]byte, 0, len(s))
				out = append(out, s[:i]...)
			}
			out = append(out, b+32) // to lower
		default:
			if out == nil {
				out = make([]byte, 0, len(s))
				out = append(out, s[:i]...)
			}
			// skip
		}
	}
	if out == nil {
		return s
	}
	return string(out)
}
