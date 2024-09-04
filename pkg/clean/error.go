package clean

import "strings"

// Error sanitizes an error message so that it can be safely logged or displayed.
func Error(err error) string {
	if err == nil {
		return "no error"
	} else if s := strings.TrimSpace(err.Error()); s == "" {
		return "unknown error"
	} else {
		// Limit error message length.
		if len(s) > LengthLimit {
			s = s[:LengthLimit]
		}

		// Remove non-printable and other potentially problematic characters.
		return strings.Map(func(r rune) rune {
			if r < 32 || r == 127 {
				return -1
			}

			switch r {
			case '`', '"':
				return '\''
			case '\\', '$', '<', '>', '{', '}':
				return '?'
			default:
				return r
			}
		}, s)
	}
}
