package txt

import "strings"

func Clip(s string, size int) string {
	if s == "" {
		return ""
	}

	s = strings.TrimSpace(s)
	runes := []rune(s)

	if len(runes) > size {
		s = string(runes[0 : size-1])
	}

	return s
}
