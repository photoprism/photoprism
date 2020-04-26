package txt

import "strings"

const (
	ClipDefault     = 160
	ClipSlug        = 80
	ClipKeyword     = 40
	ClipDescription = 16000
)

func Clip(s string, size int) string {
	s = strings.TrimSpace(s)

	if s == "" || size <= 0 {
		return ""
	}

	runes := []rune(s)

	if len(runes) > size {
		s = string(runes[0 : size-1])
	}

	return s
}
