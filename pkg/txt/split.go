package txt

import (
	"strings"
)

const (
	// EscapeRune is the default escape character for split operations.
	EscapeRune = '\\'
	// OrRune is the default OR separator for split operations.
	OrRune = '|'
	// AndRune is the default AND separator for split operations.
	AndRune = '&'
)

// SplitWithEscape splits a string by separator respecting an escape rune and optional trimming.
// If trimming, each result is trimmed and empty parts are omitted. Escape is only applied when
// the following character is escape or separator; otherwise both runes are kept.
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

		for _, r := range s {
			switch {
			case escaped:
				if r == escape || r == separator {
					upTo.WriteRune(r)
				} else {
					upTo.WriteRune(escape)
					upTo.WriteRune(r)
				}
				escaped = false
			case r == escape:
				escaped = true
			case r == separator:
				if trim {
					if t := strings.TrimSpace(upTo.String()); t != "" {
						result = append(result, t)
					}
				} else {
					result = append(result, upTo.String())
				}
				upTo.Reset()
			default:
				upTo.WriteRune(r)
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

// TrimmedSplitWithEscape splits a string with escaping and trims non-empty results.
func TrimmedSplitWithEscape(s string, separator rune, escape rune) (result []string) {
	return SplitWithEscape(s, separator, escape, true)
}

// UnTrimmedSplitWithEscape splits a string with escaping and preserves surrounding whitespace.
func UnTrimmedSplitWithEscape(s string, separator rune, escape rune) (result []string) {
	return SplitWithEscape(s, separator, escape, false)
}
