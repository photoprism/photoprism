package clean

import (
	"strings"
	"unicode"

	"github.com/photoprism/photoprism/pkg/txt"
)

// State returns the full, normalized state name.
func State(s, countryCode string) string {
	if s == "" || reject(s, txt.ClipName) {
		return Empty
	}

	// Remove whitespace from name.
	s = strings.TrimSpace(s)

	// Empty?
	if s == "" || s == txt.UnknownStateCode {
		// State doesn't have a name.
		return ""
	}

	// Remove non-printable and other potentially problematic characters.
	s = strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			return -1
		}

		switch r {
		case '~', '\\', ':', '|', '"', '?', '*', '<', '>', '{', '}':
			return -1
		default:
			return r
		}
	}, s)

	// Normalize country code.
	countryCode = strings.ToLower(strings.TrimSpace(countryCode))

	// Is the name an abbreviation that should be normalized?
	if states, found := txt.StatesByCountry[countryCode]; !found {
		// Unknown country.
	} else if normalized, found := states[s]; !found {
		// Unknown abbreviation.
	} else if normalized != "" {
		// Yes, use normalized name.
		s = normalized
	}

	// Return normalized state name.
	return s
}
