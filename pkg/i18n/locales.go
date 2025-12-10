package i18n

import (
	"strings"

	"github.com/leonelquinteros/gotext"
)

// Locale represents a language/region tag (e.g., "en", "pt_BR").
type Locale string

const (
	// German locale.
	German Locale = "de"
	// English locale.
	English Locale = "en"
	// Spanish locale.
	Spanish Locale = "es"
	// French locale.
	French Locale = "fr"
	// Dutch locale.
	Dutch Locale = "nl"
	// Polish locale.
	Polish Locale = "pl"
	// Portuguese locale.
	Portuguese Locale = "pt"
	// BrazilianPortuguese locale.
	BrazilianPortuguese Locale = "pt_BR"
	// Russian locale.
	Russian Locale = "ru"
	// ChineseSimplified locale.
	ChineseSimplified Locale = "zh"
	// ChineseTraditional locale.
	ChineseTraditional Locale = "zh_TW"
	// Default locale used when none is supplied.
	Default = English
)

var localeDir = "../../assets/locales"
var locale = Default

// SetDir sets the path to the locales directory.
func SetDir(dir string) {
	localeDir = dir
}

// SetLocale sets the current locale.
func SetLocale(loc string) {
	switch len(loc) {
	case 2:
		loc = strings.ToLower(loc[:2])
		locale = Locale(loc)
	case 5:
		loc = strings.ToLower(loc[:2]) + "_" + strings.ToUpper(loc[3:5])
		locale = Locale(loc)
	default:
		locale = Default
	}

	gotext.Configure(localeDir, string(locale), "default")
}

// Locale returns the string value of the locale.
func (l Locale) Locale() string {
	return string(l)
}
