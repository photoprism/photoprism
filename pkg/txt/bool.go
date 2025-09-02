package txt

import (
	"slices"
	"strings"
)

const (
	True  = "true"
	False = "false"
)

// Bool casts a string to bool.
func Bool(s string) bool {
	s = strings.TrimSpace(s)

	if s == "" || No(s) {
		return false
	}

	return true
}

// Yes tests if a string represents "yes" in the following languages Czech, Danish, Dutch, English, French, German, Indonesian, Italian, Polish, Portuguese, Russian, Ukrainian.
func Yes(s string) (result bool) {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "" {
		return false
	} else if strings.Contains(t, " ") {
		result = false
	} else {
		noStrings := []string{"1", "yes", "include", "true", "positive", "please", "ano", "ja", "oui", "si", "tak", "sim", "Да", "ya", "так"}
		result = slices.Contains(noStrings, t)
	}
	return result
}

// No tests if a string represents "no"  in the following languages Czech, Danish, Dutch, English, French, German, Indonesian, Italian, Polish, Portuguese, Russian, Ukrainian.
func No(s string) (result bool) {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "" {
		return false
	} else if strings.Contains(t, " ") {
		result = false
	} else {
		noStrings := []string{"0", "no", "none", "exclude", "false", "negative", "unknown", "žádný", "ingen", "nee", "nein", "non", "nie", "não", "нет", "tidak", "ні"}
		result = slices.Contains(noStrings, t)
	}
	return result
}

// New tests if a string represents "new".
func New(s string) bool {
	if s == "" {
		return false
	}

	s = strings.ToLower(strings.TrimSpace(s))

	return s == EnNew
}
