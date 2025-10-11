package txt

import (
	"strings"
	"unicode"

	"github.com/photoprism/photoprism/pkg/enum"
)

// New tests if a string represents "new".
func New(s string) bool {
	if s == "" {
		return false
	}

	s = strings.ToLower(strings.TrimSpace(s))

	return s == EnNew
}

// Bool casts a string to bool by treating any non-empty value that is not a
// known negative token as true.
func Bool(s string) bool {
	s = strings.TrimSpace(s)

	if s == "" || No(s) {
		return false
	}

	return true
}

// Yes reports whether s matches a supported affirmative token in the languages
// represented by enum.YesMap.
func Yes(s string) bool {
	return matchEnumToken(enum.YesMap, s)
}

// No reports whether s matches a supported negative token in the languages
// represented by enum.NoMap.
func No(s string) bool {
	return matchEnumToken(enum.NoMap, s)
}

// matchEnumToken normalizes s and checks whether it exists in tokens.
func matchEnumToken(tokens map[string]struct{}, s string) bool {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "" {
		return false
	}
	if strings.IndexFunc(t, unicode.IsSpace) >= 0 {
		return false
	}

	_, ok := tokens[t]
	return ok
}
