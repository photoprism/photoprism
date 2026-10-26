package txt

import (
	"unicode"
	"unicode/utf8"
)

// UpperFirst returns the string with the first character converted to uppercase.
// A string that starts with an invalid UTF-8 byte is returned unchanged.
func UpperFirst(str string) string {
	r, size := utf8.DecodeRuneInString(str)

	if r == utf8.RuneError && size <= 1 {
		return str
	}

	if upper := unicode.ToUpper(r); upper != r {
		return string(upper) + str[size:]
	}

	return str
}
