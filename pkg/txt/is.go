package txt

import "unicode"

// Is reports whether the all string runes are in the specified range.
func Is(rangeTab *unicode.RangeTable, s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !unicode.Is(rangeTab, r) {
			return false
		}
	}

	return true
}

// IsASCII tests if the string only contains ascii runes.
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// IsLatin reports whether the string only contains latin letters.
func IsLatin(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !unicode.Is(unicode.Latin, r) {
			return false
		}
	}

	return true
}
