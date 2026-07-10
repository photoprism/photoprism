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

// PosixLocale returns the normalized locale string in POSIX format with underscore, or the default locale otherwise.
// See https://en.wikipedia.org/wiki/Locale_(computer_software) for details.
func PosixLocale(locale, defaultLocale string) string {
	return Locale(locale, defaultLocale)
}

// WebLocale returns a normalized locale string in BCP 47 format with a dash, or the default locale otherwise.
// See https://en.wikipedia.org/wiki/Locale_(computer_software) for details.
func WebLocale(locale, defaultLocale string) string {
	if locale == "" {
		return defaultLocale
	}

	locale, _, _ = strings.Cut(strings.Replace(locale, "_", "-", 1), ".")

	if l := len(locale); l == 2 {
		return strings.ToLower(locale)
	} else if l == 5 && locale[2] == '-' {
		return strings.ToLower(locale[:2]) + "-" + strings.ToUpper(locale[3:])
	}

	return defaultLocale
}

// rtlLocales contains the ISO 639 primary language subtags rendered right-to-left.
var rtlLocales = map[string]bool{
	"ar": true, // Arabic
	"fa": true, // Persian
	"he": true, // Hebrew
	"ku": true, // Kurdish (Sorani)
}

// TextDir returns the user interface text direction ("rtl" or "ltr") for the specified locale,
// falling back to defaultLocale when locale is empty or invalid, and to "ltr" otherwise.
func TextDir(locale, defaultLocale string) string {
	for _, l := range []string{locale, defaultLocale} {
		// WebLocale normalizes the language subtag to lowercase, so no extra ToLower is needed.
		if l = WebLocale(l, ""); l == "" {
			continue
		} else if lang, _, _ := strings.Cut(l, "-"); rtlLocales[lang] {
			return "rtl"
		} else {
			return "ltr"
		}
	}

	return "ltr"
}
