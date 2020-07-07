package i18n

import "strings"

type Language string

type LanguageMap map[Language]MessageMap

const (
	German     Language = "de"
	English    Language = "en"
	Spanish    Language = "es"
	French     Language = "fr"
	Dutch      Language = "nl"
	Portuguese Language = "pt"
	Russian    Language = "ru"
	Chinese    Language = "zh"
	Default             = English
)

var Languages = LanguageMap{
	German:     MsgGerman,
	English:    MsgEnglish,
	Spanish:    MsgSpanish,
	French:     MsgFrench,
	Dutch:      MsgDutch,
	Portuguese: MsgPortuguese,
	Russian:    MsgRussian,
	Chinese:    MsgChinese,
}

var Lang = Default

func SetLang(s string) {
	if len(s) != 2 {
		Lang = Default
	} else {
		s = strings.ToLower(s)
		Lang = Language(s)
	}
}
