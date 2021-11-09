package txt

import "strings"

// NormalizeName sanitizes and capitalizes names.
func NormalizeName(name string) string {
	if name == "" {
		return ""
	}

	// Remove double quotes and other special characters.
	name = strings.Map(func(r rune) rune {
		switch r {
		case '"', '`', '~', '\\', '/', '*', '%', '&', '|', '+', '=', '$', '@', '!', '?', ':', ';', '<', '>', '{', '}':
			return -1
		}
		return r
	}, name)

	name = strings.TrimSpace(name)

	if name == "" {
		return ""
	}

	// Shorten and capitalize.
	return Clip(Title(name), ClipDefault)
}

// NormalizeState returns the full, normalized state name.
func NormalizeState(s string) string {
	s = strings.TrimSpace(s)

	if s == "" || s == UnknownStateCode {
		return ""
	}

	if expanded, ok := States[s]; ok {
		return expanded
	}

	return s
}

// NormalizeQuery replaces search operator with default symbols.
func NormalizeQuery(s string) string {
	s = strings.ToLower(Clip(s, ClipQuery))
	s = strings.ReplaceAll(s, Spaced(EnOr), Or)
	s = strings.ReplaceAll(s, Spaced(EnAnd), And)
	s = strings.ReplaceAll(s, Spaced(EnWith), And)
	s = strings.ReplaceAll(s, Spaced(EnIn), And)
	s = strings.ReplaceAll(s, Spaced(EnAt), And)
	s = strings.ReplaceAll(s, SpacedPlus, And)
	s = strings.ReplaceAll(s, "%", "*")
	return strings.Trim(s, "+&|_-=!@$%^(){}\\<>,.;: ")
}
