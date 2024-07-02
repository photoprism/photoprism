package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocale(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Locale("", ""))
		assert.Equal(t, "de", Locale("", "de"))
		assert.Equal(t, "und", Locale("", "und"))
	})
	t.Run("Language", func(t *testing.T) {
		assert.Equal(t, "de", Locale("de", ""))
		assert.Equal(t, "", Locale("und", ""))
		assert.Equal(t, "de", Locale("und", "de"))
		assert.Equal(t, "cs", Locale("cs", "und"))
	})
	t.Run("Territory", func(t *testing.T) {
		assert.Equal(t, "cs_CZ", Locale("cs_CZ", ""))
		assert.Equal(t, "cs_CZ", Locale("cs-CZ", ""))
		assert.Equal(t, "cs_CZ", Locale("cs_cz", ""))
		assert.Equal(t, "cs_CZ", Locale("cs-cz", ""))
		assert.Equal(t, "cs_CZ", Locale("Cs_cz", ""))
		assert.Equal(t, "cs_CZ", Locale("Cs-cz", ""))
		assert.Equal(t, "cs_CZ", Locale("cs_CZ", "und"))
		assert.Equal(t, "cs_CZ", Locale("cs-CZ", "und"))
		assert.Equal(t, "und", Locale("cs-CZX", "und"))
	})
}
