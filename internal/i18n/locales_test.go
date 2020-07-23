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
	SetLocale("PL")
	assert.Equal(t, Polish, locale)
	SetLocale("")
	assert.Equal(t, English, locale)
	assert.Equal(t, Default, locale)
}
