package clean

import "strings"

// Locale returns the normalized locale string in POSIX format with underscore, or the default locale otherwise.
// See https://en.wikipedia.org/wiki/Locale_(computer_software) for details.
func Locale(locale, defaultLocale string) string {
	if locale == "" {
		return defaultLocale
	}

	locale, _, _ = strings.Cut(strings.Replace(locale, "-", "_", 1), ".")

	if l := len(locale); l == 2 {
		return strings.ToLower(locale)
	} else if l == 5 && locale[2] == '_' {
		return strings.ToLower(locale[:2]) + "_" + strings.ToUpper(locale[3:])
	}

	return defaultLocale
}
