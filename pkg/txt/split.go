package txt

import (
	"strings"
)

const (
	EscapeRune = '\\'
	OrRune     = '|'
	AndRune    = '&'
)

// Splits a string using a separator allowing escaping that separator with escape and optionally trims the results
// If trimming then Trims each result, and doesn't return empty sets
// Only applies the escape if the following character is escape or separator
// otherwise it keeps both runes
func SplitWithEscape(s string, separator rune, escape rune, trim bool) (result []string) {
	if s == "" {
		return []string{}
	}

	result = []string{}

	if !strings.ContainsRune(s, separator) {
		result = append(result, s)
	} else {
		escaped := false
		var upTo strings.Builder

		for _, rune := range s {
			if escaped {
				if rune == escape || rune == separator {
					upTo.WriteRune(rune)
				} else {
					upTo.WriteRune(escape)
					upTo.WriteRune(rune)
				}
				escaped = false
			} else if rune == escape {
				escaped = true
			} else if rune == separator {
				if trim {
					if t := strings.TrimSpace(upTo.String()); t != "" {
						result = append(result, t)
					}
				} else {
					result = append(result, upTo.String())
				}
				upTo.Reset()
			} else {
				upTo.WriteRune(rune)
			}
		}
		if trim {
			if t := strings.TrimSpace(upTo.String()); t != "" {
				result = append(result, t)
			}
		} else {
			result = append(result, upTo.String())
		}
	}
	return result
}

// Splits a string using a separator allowing escaping that separator with escape
// Trims each result, and doesn't return empty sets
// Only applies the escape if the following character is escape or separator
// otherwise it keeps both runes
func TrimmedSplitWithEscape(s string, separator rune, escape rune) (result []string) {
	return SplitWithEscape(s, separator, escape, true)
}

// Splits a string using a separator allowing escaping that separator with escape
// Does not trim each result, and will return sets containing just white space if included in string
// Only applies the escape if the following character is escape or separator
// otherwise it keeps both runes
func UnTrimmedSplitWithEscape(s string, separator rune, escape rune) (result []string) {
	return SplitWithEscape(s, separator, escape, false)
}
