package entity

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// trimMakePrefix removes a leading make name from a camera or lens model name.
// The make is removed only as a whole word, so that e.g. "Nikonos V" or "Helios-44-2" stay intact.
func trimMakePrefix(modelName, makeName string) string {
	if makeName == "" || !strings.HasPrefix(modelName, makeName) {
		return modelName
	}

	rest := modelName[len(makeName):]

	if rest == "" {
		return ""
	} else if r, _ := utf8.DecodeRuneInString(rest); !unicode.IsSpace(r) {
		return modelName
	}

	return strings.TrimSpace(rest)
}
