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

// IsNumeric tests if the string starts and ends with an ascii number character.
func IsNumeric(s string) bool {
	if s == "" {
		return false
	}

	if s[0] < '0' || s[0] > '9' {
		return false
	}

	if l := len(s); l < 2 {
		return true
	} else if r := s[l-1]; r < '0' || r > '9' {
		return false
	}

	return true
}

// IsNumeral tests if the string is a word describing a number.
func IsNumeral(s string) bool {
	if s == "" {
		return false
	}

	if s[0] < '0' || s[0] > '9' {
		return false
	}

	if l := len(s); l < 2 {
		return false
	} else if r := s[l-1]; r >= 'A' {
		return true
	}

	return false
}

// IsNumber tests if the string only contains ascii number characters.
func IsNumber(s string) bool {
	if s == "" {
		return false
	}

	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}

// IsDateNumber tests if the string only contains numeric characters, common delimiters like "-" and "_".
func IsDateNumber(s string) bool {
	if s == "" {
		return false
	}

	for i := 0; i < len(s); i++ {
		if (s[i] < '0' || s[i] > '9') && s[i] != '_' && s[i] != '-' {
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
