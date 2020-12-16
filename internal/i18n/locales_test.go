package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetLocale(t *testing.T) {
	assert.Equal(t, English, locale)
	SetLocale("D")
	assert.Equal(t, English, locale)
	SetLocale("De")
	assert.Equal(t, German, locale)
	SetLocale("de_AT")
	assert.Equal(t, Locale("de_AT"), locale)
	SetLocale("pt")
	assert.Equal(t, Portuguese, locale)
	SetLocale("pt_br")
	assert.Equal(t, BrazilianPortuguese, locale)
	SetLocale("zh")
	assert.Equal(t, ChineseSimplified, locale)
	SetLocale("zh_TW")
	assert.Equal(t, ChineseTraditional, locale)
	SetLocale("ru")
	assert.Equal(t, Russian, locale)
	SetLocale("fr")
	assert.Equal(t, French, locale)
	SetLocale("nl")
	assert.Equal(t, Dutch, locale)
	SetLocale("es")
	assert.Equal(t, Spanish, locale)
	SetLocale("PL")
	assert.Equal(t, Polish, locale)
	SetLocale("")
	assert.Equal(t, English, locale)
	assert.Equal(t, Default, locale)
}
