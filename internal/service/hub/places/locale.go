package places

import (
	"github.com/photoprism/photoprism/pkg/clean"
)

// LocalLocale specifies the locale name to return results in the local language.
const LocalLocale = "local"

// DefaultLocale specifies the default places query locale.
var DefaultLocale = LocalLocale

// Locale returns the places query locale string.
func Locale(locale string) string {
	return clean.WebLocale(locale, DefaultLocale)
}
