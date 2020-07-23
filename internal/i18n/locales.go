package i18n

import (
	"strings"

	"github.com/leonelquinteros/gotext"
)

type Locale string

const (
	German     Locale = "de"
	English    Locale = "en"
	Spanish    Locale = "es"
	French     Locale = "fr"
	Dutch      Locale = "nl"
	Polish     Locale = "pl"
	Portuguese Locale = "pt"
	Russian    Locale = "ru"
	Chinese    Locale = "zh"
	Default           = English
)

var localeDir = "../../assets/locales"
var locale = Default

func SetDir(dir string) {
	localeDir = dir
}

func SetLocale(loc string) {
	if len(loc) != 2 {
		locale = Default
	} else {
		loc = strings.ToLower(loc)
		locale = Locale(loc)
	}

	gotext.Configure(localeDir, string(locale), "default")
}
