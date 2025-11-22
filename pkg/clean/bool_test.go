package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	t.Run("TrueTokens", func(t *testing.T) {
		assert.True(t, Bool("t"))
		assert.True(t, Bool("Yes"))
	})
	t.Run("FalseTokens", func(t *testing.T) {
		assert.False(t, Bool("F"))
		assert.False(t, Bool("no"))
	})
}

func TestYes(t *testing.T) {
	t.Run("ShortTrue", func(t *testing.T) {
		assert.True(t, Yes("t"))
	})
	t.Run("Localized", func(t *testing.T) {
		assert.True(t, Yes("oui"))
	})
}

func TestNo(t *testing.T) {
	t.Run("ShortFalse", func(t *testing.T) {
		assert.True(t, No("f"))
	})
	t.Run("Localized", func(t *testing.T) {
		assert.True(t, No("nein"))
	})
}
