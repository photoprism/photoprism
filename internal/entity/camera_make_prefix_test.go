package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimMakePrefix(t *testing.T) {
	t.Run("Space", func(t *testing.T) {
		assert.Equal(t, "EOS 7D", trimMakePrefix("Canon EOS 7D", "Canon"))
		assert.Equal(t, "D750", trimMakePrefix("NIKON  D750 ", "NIKON"))
		assert.Equal(t, "X-700", trimMakePrefix("Minolta\tX-700", "Minolta"))
	})
	t.Run("Hyphen", func(t *testing.T) {
		assert.Equal(t, "H815", trimMakePrefix("LG-H815", "LG"))
		assert.Equal(t, "SM-G900A", trimMakePrefix("SAMSUNG-SM-G900A", "SAMSUNG"))
		assert.Equal(t, "44-2 58mm f/2", trimMakePrefix("Helios-44-2 58mm f/2", "Helios"))
		assert.Equal(t, "Mat 124G", trimMakePrefix("Yashica-Mat 124G", "Yashica"))
	})
	t.Run("Underscore", func(t *testing.T) {
		// Slugs keep underscores, so they must stay in the model to keep existing slugs unchanged.
		assert.Equal(t, "_AI2302", trimMakePrefix("ASUS_AI2302", "ASUS"))
		assert.Equal(t, "_One", trimMakePrefix("HTC_One", "HTC"))
	})
	t.Run("Digit", func(t *testing.T) {
		assert.Equal(t, "5Z100", trimMakePrefix("Apple5Z100", "Apple"))
	})
	t.Run("SameAsMake", func(t *testing.T) {
		assert.Equal(t, "", trimMakePrefix("Zenit", "Zenit"))
		assert.Equal(t, "", trimMakePrefix("Zenit -", "Zenit"))
	})
	t.Run("PartOfWord", func(t *testing.T) {
		assert.Equal(t, "Nikonos V", trimMakePrefix("Nikonos V", "Nikon"))
		assert.Equal(t, "Canonet QL17 GIII", trimMakePrefix("Canonet QL17 GIII", "Canon"))
		assert.Equal(t, "Rolleiflex 2.8F", trimMakePrefix("Rolleiflex 2.8F", "Rollei"))
		assert.Equal(t, "Leicaflex SL", trimMakePrefix("Leicaflex SL", "Leica"))
	})
	t.Run("NoPrefix", func(t *testing.T) {
		assert.Equal(t, "EOS 7D", trimMakePrefix("EOS 7D", "Canon"))
		assert.Equal(t, "canon EOS 7D", trimMakePrefix("canon EOS 7D", "Canon"))
	})
	t.Run("EmptyMake", func(t *testing.T) {
		assert.Equal(t, "EOS 7D", trimMakePrefix("EOS 7D", ""))
	})
	t.Run("EmptyModel", func(t *testing.T) {
		assert.Equal(t, "", trimMakePrefix("", "Canon"))
	})
}
