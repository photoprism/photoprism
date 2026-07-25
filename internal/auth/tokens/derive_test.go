package tokens

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDerive(t *testing.T) {
	key := make([]byte, KeyLen)
	key[0] = 0x42
	other := make([]byte, KeyLen)
	other[0] = 0x43
	t.Run("Stable", func(t *testing.T) {
		// Must not change between calls or restarts, or every cached preview URL would break.
		assert.Equal(t, Derive(key, PurposePreview), Derive(key, PurposePreview))
	})
	t.Run("HexOfExpectedLength", func(t *testing.T) {
		token := Derive(key, PurposePreview)
		assert.Len(t, token, deriveLen*2)
		assert.Equal(t, strings.ToLower(token), token)
	})
	t.Run("DiffersByKey", func(t *testing.T) {
		assert.NotEqual(t, Derive(key, PurposePreview), Derive(other, PurposePreview))
	})
	t.Run("DiffersByPurpose", func(t *testing.T) {
		assert.NotEqual(t, Derive(key, PurposePreview), Derive(key, "download"))
	})
	t.Run("DoesNotRevealKey", func(t *testing.T) {
		// The token is published in URLs, so it must not contain the key it was derived from.
		assert.NotContains(t, Derive(key, PurposePreview), hexKey(key))
	})
	t.Run("ShortKey", func(t *testing.T) {
		assert.Empty(t, Derive(make([]byte, KeyLen-1), PurposePreview))
	})
	t.Run("NilKey", func(t *testing.T) {
		assert.Empty(t, Derive(nil, PurposePreview))
	})
	t.Run("EmptyPurpose", func(t *testing.T) {
		assert.Empty(t, Derive(key, ""))
	})
}

// hexKey renders a key as a lowercase hex string for use in assertions.
func hexKey(key []byte) string {
	var b strings.Builder

	for _, c := range key {
		b.WriteString(string("0123456789abcdef"[c>>4]))
		b.WriteString(string("0123456789abcdef"[c&0x0f]))
	}

	return b.String()
}
