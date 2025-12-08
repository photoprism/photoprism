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

func TestPosixLocale(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", PosixLocale("", ""))
		assert.Equal(t, "de", PosixLocale("", "de"))
		assert.Equal(t, "und", PosixLocale("", "und"))
	})
	t.Run("Language", func(t *testing.T) {
		assert.Equal(t, "de", PosixLocale("de", ""))
		assert.Equal(t, "", PosixLocale("und", ""))
		assert.Equal(t, "de", PosixLocale("und", "de"))
		assert.Equal(t, "cs", PosixLocale("cs", "und"))
	})
	t.Run("Local", func(t *testing.T) {
		assert.Equal(t, "local", PosixLocale("", "local"))
		assert.Equal(t, "Local", PosixLocale("", "Local"))
		assert.Equal(t, "", PosixLocale("local", ""))
		assert.Equal(t, "", PosixLocale("Local", ""))
		assert.Equal(t, "local", PosixLocale("local", "local"))
	})
	t.Run("Territory", func(t *testing.T) {
		assert.Equal(t, "cs_CZ", PosixLocale("cs_CZ", ""))
		assert.Equal(t, "cs_CZ", PosixLocale("cs-CZ", ""))
		assert.Equal(t, "cs_CZ", PosixLocale("cs_cz", ""))
		assert.Equal(t, "cs_CZ", PosixLocale("cs-cz", ""))
		assert.Equal(t, "cs_CZ", PosixLocale("Cs_cz", ""))
		assert.Equal(t, "cs_CZ", PosixLocale("Cs-cz", ""))
		assert.Equal(t, "cs_CZ", PosixLocale("cs_CZ", "und"))
		assert.Equal(t, "cs_CZ", PosixLocale("cs-CZ", "und"))
		assert.Equal(t, "und", PosixLocale("cs-CZX", "und"))
	})
}

func TestWebLocale(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", WebLocale("", ""))
		assert.Equal(t, "de", WebLocale("", "de"))
		assert.Equal(t, "und", WebLocale("", "und"))
	})
	t.Run("Language", func(t *testing.T) {
		assert.Equal(t, "de", WebLocale("de", ""))
		assert.Equal(t, "", WebLocale("und", ""))
		assert.Equal(t, "de", WebLocale("und", "de"))
		assert.Equal(t, "cs", WebLocale("cs", "und"))
	})
	t.Run("Local", func(t *testing.T) {
		assert.Equal(t, "local", WebLocale("", "local"))
		assert.Equal(t, "Local", WebLocale("", "Local"))
		assert.Equal(t, "", WebLocale("local", ""))
		assert.Equal(t, "", WebLocale("Local", ""))
		assert.Equal(t, "local", WebLocale("local", "local"))
	})
	t.Run("Territory", func(t *testing.T) {
		assert.Equal(t, "cs-CZ", WebLocale("cs-CZ", ""))
		assert.Equal(t, "cs-CZ", WebLocale("cs_CZ", ""))
		assert.Equal(t, "cs-CZ", WebLocale("cs-cz", ""))
		assert.Equal(t, "cs-CZ", WebLocale("cs_cz", ""))
		assert.Equal(t, "cs-CZ", WebLocale("Cs-cz", ""))
		assert.Equal(t, "cs-CZ", WebLocale("Cs_cz", ""))
		assert.Equal(t, "cs-CZ", WebLocale("cs-CZ", "und"))
		assert.Equal(t, "cs-CZ", WebLocale("cs_CZ", "und"))
		assert.Equal(t, "und", WebLocale("cs_CZX", "und"))
	})
}
