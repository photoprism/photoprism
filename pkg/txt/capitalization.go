package txt

import (
	"strings"
	"unicode"
)

// isSeparator reports whether the rune could mark a word boundary.
func isSeparator(r rune) bool {
	// ASCII alphanumerics and underscore are not separators
	if r <= 0x7F {
		switch {
		case '0' <= r && r <= '9':
			return false
		case 'a' <= r && r <= 'z':
			return false
		case 'A' <= r && r <= 'Z':
			return false
		case r == '_', r == '\'':
			return false
		}
		return true
	}
	// Letters and digits are not separators
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	// Otherwise, all we can do for now is treat spaces as separators.
	return unicode.IsSpace(r)
}

// UcFirst returns the string with the first character converted to uppercase.
func UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// Title returns the string with the first characters of each word converted to uppercase.
func Title(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.Trim(s, "/ -")

	if s == "" {
		return ""
	}

	blocks := strings.Split(s, "/")
	result := make([]string, 0, len(blocks))

	for _, block := range blocks {
		words := strings.Fields(block)

		if len(words) == 0 {
			continue
		}

		for i, w := range words {
			search := strings.ToLower(strings.Trim(w, ":.,;!?"))

			if match, ok := SpecialWords[search]; ok {
				words[i] = strings.Replace(strings.ToLower(w), search, match, 1)
			} else if i > 0 && SmallWords[search] {
				words[i] = strings.ToLower(w)
			} else {
				prev := ' '
				words[i] = strings.Map(
					func(r rune) rune {
						if isSeparator(prev) {
							prev = r
							return unicode.ToTitle(r)
						}
						prev = r
						return r
					},
					w)
			}
		}

		result = append(result, strings.Join(words, " "))
	}

	return strings.Join(result, " / ")
}
