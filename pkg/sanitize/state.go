package sanitize

import (
	"strings"
	"unicode"

	"github.com/photoprism/photoprism/pkg/txt"
)

// State returns the full, normalized state name.
func State(stateName, countryCode string) string {
	// Remove whitespace from name.
	stateName = strings.TrimSpace(stateName)

	// Empty?
	if stateName == "" || stateName == txt.UnknownStateCode {
		// State doesn't have a name.
		return ""
	}

	// Remove non-printable and other potentially problematic characters.
	stateName = strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			return -1
		}

		switch r {
		case '~', '\\', ':', '|', '"', '?', '*', '<', '>', '{', '}':
			return -1
		default:
			return r
		}
	}, stateName)

	// Normalize country code.
	countryCode = strings.ToLower(strings.TrimSpace(countryCode))

	// Is the name an abbreviation that should be normalized?
	if states, found := txt.StatesByCountry[countryCode]; !found {
		// Unknown country.
	} else if normalized, found := states[stateName]; !found {
		// Unknown abbreviation.
	} else if normalized != "" {
		// Yes, use normalized name.
		stateName = normalized
	}

	// Return normalized state name.
	return stateName
}
