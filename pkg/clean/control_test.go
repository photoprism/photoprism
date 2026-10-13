package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripControl(t *testing.T) {
	t.Run("Unchanged", func(t *testing.T) {
		assert.Equal(t, "", StripControl(""))
		assert.Equal(t, "Europe/Berlin", StripControl("Europe/Berlin"))
		assert.Equal(t, "/Fotos/Älter 2024 幸福", StripControl("/Fotos/Älter 2024 幸福"))
		assert.Zero(t, testing.AllocsPerRun(10, func() { _ = StripControl("Europe/Berlin") }))
	})
	t.Run("Whitespace", func(t *testing.T) {
		assert.Equal(t, "a b c d e f", StripControl("a\tb\nc\vd\fe\rf"))
	})
	t.Run("Removed", func(t *testing.T) {
		assert.Equal(t, "ab[31mcd", StripControl("a\x00b\x1b[31mc\x7fd"))
		assert.Equal(t, "", StripControl("\x00\x01\x1f\x7f"))
		assert.Equal(t, "Ä幸", StripControl("Ä\x00幸"))
	})
	t.Run("InvalidUTF8Kept", func(t *testing.T) {
		assert.Equal(t, "a\xffb", StripControl("a\xff\x00b"))
	})
}
