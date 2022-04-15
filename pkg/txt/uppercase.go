package txt

import (
	"unicode"
)

// UpperFirst returns the string with the first character converted to uppercase.
func UpperFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}
