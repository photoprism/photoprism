package clean

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// State returns the normalized state name clipped to at most txt.ClipName
// runes; injection patterns (`${`, `ldap://`) drop the value entirely.
func State(s, countryCode string) string {
	if s == "" || reject(s, 0) {
		return Empty
	}

	// Remove whitespace from name.
	s = strings.TrimSpace(s)

	// Empty?
	if s == "" || s == txt.UnknownStateCode {
		// State doesn't have a name.
		return ""
	}

	// Replace lookalike spaces and remove non-printable characters except U+200C and U+200D.
	s = strings.Map(func(r rune) rune {
		if unsafeSpaceRune(r) {
			return unsafeSpace
		} else if unsafeRune(r) {
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

	return txt.Clip(s, txt.ClipName)
}
