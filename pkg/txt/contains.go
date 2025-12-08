package txt

import "unicode"

// ContainsNumber returns true if string contains an ASCII digit.
func ContainsNumber(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			return true
		}
	}
	return false
}

// ContainsLetters reports whether the string only contains letters.
func ContainsLetters(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}

	return true
}

// ContainsASCIILetters reports if the string only contains ascii chars without whitespace, numbers, and punctuation marks.
func ContainsASCIILetters(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if (b < 'A' || b > 'Z') && (b < 'a' || b > 'z') {
			return false
		}
	}
	return true
}

// ContainsAlnumLower reports if the string only contains lower case ascii letters or numbers.
func ContainsAlnumLower(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if (b < '0' || b > '9') && (b < 'a' || b > 'z') {
			return false
		}
	}
	return true
}

// ContainsSymbols reports whether the string only contains symbolic characters.
func ContainsSymbols(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !unicode.IsSymbol(r) {
			return false
		}
	}

	return true
}
