package authtoken

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSign(t *testing.T) {
	key := []byte("test-signing-key-0123456789abcdef")
	t.Run("Format", func(t *testing.T) {
		token := Sign(key, "/api/v1", 1_700_000_000, url.Values{"sid": {"sess123"}}, "")
		assert.True(t, strings.HasPrefix(token, Prefix))
		assert.NotContains(t, token, "=") // base64url without padding
	})
	t.Run("Deterministic", func(t *testing.T) {
		a := Sign(key, "/api/v1", 1_700_000_000, url.Values{"sid": {"sess123"}}, "")
		b := Sign(key, "/api/v1", 1_700_000_000, url.Values{"sid": {"sess123"}}, "")
		assert.Equal(t, a, b)
	})
	t.Run("PathChangesSignature", func(t *testing.T) {
		a := Sign(key, "/api/v1", 1_700_000_000, url.Values{"sid": {"sess123"}}, "")
		b := Sign(key, "/api/v2", 1_700_000_000, url.Values{"sid": {"sess123"}}, "")
		assert.NotEqual(t, a, b)
	})
	t.Run("ParamsChangeSignature", func(t *testing.T) {
		a := Sign(key, "/api/v1", 1_700_000_000, url.Values{"sid": {"sess123"}}, "")
		b := Sign(key, "/api/v1", 1_700_000_000, url.Values{"sid": {"other"}}, "")
		assert.NotEqual(t, a, b)
	})
	t.Run("ClientIPChangesSignature", func(t *testing.T) {
		a := Sign(key, "/api/v1", 1_700_000_000, url.Values{"sid": {"sess123"}}, "")
		b := Sign(key, "/api/v1", 1_700_000_000, url.Values{"sid": {"sess123"}}, "203.0.113.7")
		assert.NotEqual(t, a, b)
	})
}

func TestSigningData(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SigningData(url.Values{}))
	})
	t.Run("SortedAndExcludesTokenExpires", func(t *testing.T) {
		params := url.Values{"b": {"2"}, "a": {"1"}, "token": {"x"}, "expires": {"9"}}
		assert.Equal(t, "a=1&b=2", SigningData(params))
	})
	t.Run("SingleParam", func(t *testing.T) {
		assert.Equal(t, "sid=sess123", SigningData(url.Values{"sid": {"sess123"}}))
	})
}
