package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimMakePrefix(t *testing.T) {
	t.Run("WholeWord", func(t *testing.T) {
		assert.Equal(t, "EOS 7D", trimMakePrefix("Canon EOS 7D", "Canon"))
		assert.Equal(t, "D750", trimMakePrefix("NIKON  D750 ", "NIKON"))
		assert.Equal(t, "X-700", trimMakePrefix("Minolta\tX-700", "Minolta"))
	})
	t.Run("SameAsMake", func(t *testing.T) {
		assert.Equal(t, "", trimMakePrefix("Zenit", "Zenit"))
	})
	t.Run("PartOfWord", func(t *testing.T) {
		assert.Equal(t, "Nikonos V", trimMakePrefix("Nikonos V", "Nikon"))
		assert.Equal(t, "Canonet QL17 GIII", trimMakePrefix("Canonet QL17 GIII", "Canon"))
		assert.Equal(t, "Rolleiflex 2.8F", trimMakePrefix("Rolleiflex 2.8F", "Rollei"))
		assert.Equal(t, "Helios-44-2 58mm f/2", trimMakePrefix("Helios-44-2 58mm f/2", "Helios"))
		assert.Equal(t, "Yashica-Mat 124G", trimMakePrefix("Yashica-Mat 124G", "Yashica"))
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
