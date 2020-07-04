package i18n

import "strings"

type Language string

type LanguageMap map[Language]MessageMap

const (
	English Language = "en"
	Dutch   Language = "nl"
	French  Language = "fr"
	German  Language = "de"
	Russian Language = "ru"
	Default          = English
)

var Languages = LanguageMap{
	English: MsgEnglish,
	Dutch:   MsgDutch,
	French:  MsgFrench,
	German:  MsgGerman,
	Russian: MsgRussian,
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
