package entity

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// trimMakePrefix removes a leading make name from a camera or lens model name, e.g. "LG" from "LG-H815".
// The make is kept if a letter follows it, as in "Nikonos V", so that it is not cut off from a longer word.
// Only spaces and hyphens are trimmed after it, since slugs keep underscores and must not change for known devices.
func trimMakePrefix(modelName, makeName string) string {
	if makeName == "" || !strings.HasPrefix(modelName, makeName) {
		return modelName
	}

	rest := modelName[len(makeName):]

	if r, _ := utf8.DecodeRuneInString(rest); rest != "" && unicode.IsLetter(r) {
		return modelName
	}

	return strings.TrimRightFunc(strings.TrimLeftFunc(rest, func(r rune) bool {
		return unicode.IsSpace(r) || r == '-'
	}), unicode.IsSpace)
}
