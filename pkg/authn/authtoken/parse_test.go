package authtoken

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerify(t *testing.T) {
	key := []byte("test-signing-key-0123456789abcdef")
	path := "/api/v1"
	exp := int64(1_700_000_000)
	now := exp - 3600
	params := url.Values{"sid": {"sess123"}}
	t.Run("Valid", func(t *testing.T) {
		token := Sign(key, path, exp, params, "")
		assert.NoError(t, Verify(key, path, exp, params, "", token, now))
	})
	t.Run("Expired", func(t *testing.T) {
		token := Sign(key, path, exp, params, "")
		assert.ErrorIs(t, Verify(key, path, exp, params, "", token, exp+1), ErrExpired)
	})
	t.Run("WrongKey", func(t *testing.T) {
		token := Sign(key, path, exp, params, "")
		assert.ErrorIs(t, Verify([]byte("another-key-that-differs-abcdef012"), path, exp, params, "", token, now), ErrSignature)
	})
	t.Run("TamperedPath", func(t *testing.T) {
		token := Sign(key, path, exp, params, "")
		assert.ErrorIs(t, Verify(key, "/api/v2", exp, params, "", token, now), ErrSignature)
	})
	t.Run("TamperedParams", func(t *testing.T) {
		token := Sign(key, path, exp, params, "")
		assert.ErrorIs(t, Verify(key, path, exp, url.Values{"sid": {"attacker"}}, "", token, now), ErrSignature)
	})
	t.Run("TamperedExpires", func(t *testing.T) {
		token := Sign(key, path, exp, params, "")
		// A client that extends expires without re-signing must fail (expires is in the message).
		assert.ErrorIs(t, Verify(key, path, exp+10_000, params, "", token, now), ErrSignature)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.ErrorIs(t, Verify(key, path, exp, params, "", "", now), ErrMalformed)
	})
}

func TestValid(t *testing.T) {
	key := []byte("test-signing-key-0123456789abcdef")
	path := "/api/v1"
	exp := int64(1_700_000_000)
	params := url.Values{"sid": {"sess123"}}
	t.Run("True", func(t *testing.T) {
		assert.True(t, Valid(key, path, exp, params, "", Sign(key, path, exp, params, ""), exp-1))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, Valid(key, path, exp, params, "", Sign(key, path, exp, params, ""), exp+1))
	})
}
