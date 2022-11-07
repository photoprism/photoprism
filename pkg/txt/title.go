package txt

import (
	"strings"
	"unicode"
)

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
			} else if i > 0 && SmallWords[search] == true {
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
