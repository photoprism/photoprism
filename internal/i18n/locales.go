package i18n

import (
	"strings"

	"github.com/leonelquinteros/gotext"
)

type Locale string

const (
	German              Locale = "de"
	English             Locale = "en"
	Spanish             Locale = "es"
	French              Locale = "fr"
	Dutch               Locale = "nl"
	Polish              Locale = "pl"
	Portuguese          Locale = "pt"
	BrazilianPortuguese Locale = "pt_BR"
	Russian             Locale = "ru"
	ChineseSimplified   Locale = "zh"
	ChineseTraditional  Locale = "zh_TW"
	Default                    = English
)

var localeDir = "../../assets/locales"
var locale = Default

func SetDir(dir string) {
	localeDir = dir
}

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

func (l Locale) Locale() string {
	return string(l)
}
