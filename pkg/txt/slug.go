package txt

import "strings"

// SlugToTitle converts a slug back to a title
func SlugToTitle(s string) string {
	if s == "" {
		return ""
	}

	return Title(strings.Join(Words(s), " "))
}
