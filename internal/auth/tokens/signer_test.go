package tokens

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/authn/authtoken"
)

func TestSigner_Configured(t *testing.T) {
	t.Run("NilSigner", func(t *testing.T) {
		var s *Signer
		assert.False(t, s.Configured())
	})
	t.Run("ShortKey", func(t *testing.T) {
		assert.False(t, (&Signer{Key: make([]byte, KeyLen-1)}).Configured())
	})
	t.Run("FullKey", func(t *testing.T) {
		assert.True(t, (&Signer{Key: make([]byte, KeyLen)}).Configured())
	})
}

func TestSigner_Sign(t *testing.T) {
	params := url.Values{"sid": {"sess123"}}

	t.Run("Unconfigured", func(t *testing.T) {
		assert.Empty(t, (&Signer{}).Sign(1_700_000_000, params))
	})
	t.Run("Configured", func(t *testing.T) {
		s := &Signer{Key: make([]byte, KeyLen), SignaturePath: "/api/v1"}
		token := s.Sign(1_700_000_000, params)
		assert.True(t, strings.HasPrefix(token, authtoken.Prefix))
		// Signing again yields the identical value: the message is deterministic (no per-call nonce).
		assert.Equal(t, token, s.Sign(1_700_000_000, params))
	})
}

func TestSigner_Valid(t *testing.T) {
	s := &Signer{Key: make([]byte, KeyLen), SignaturePath: "/api/v1"}
	params := url.Values{"sid": {"sess123"}}
	expires := time.Now().Add(time.Hour).Unix()
	token := s.Sign(expires, params)

	t.Run("Valid", func(t *testing.T) {
		assert.True(t, s.Valid(expires, params, token))
	})
	t.Run("Expired", func(t *testing.T) {
		past := time.Now().Add(-time.Hour).Unix()
		assert.False(t, s.Valid(past, params, s.Sign(past, params)))
	})
	t.Run("WrongParams", func(t *testing.T) {
		assert.False(t, s.Valid(expires, url.Values{"sid": {"attacker"}}, token))
	})
	t.Run("EmptyToken", func(t *testing.T) {
		assert.False(t, s.Valid(expires, params, ""))
	})
	t.Run("OversizedToken", func(t *testing.T) {
		assert.False(t, s.Valid(expires, params, strings.Repeat("x", maxTokenLen+1)))
	})
	t.Run("Unconfigured", func(t *testing.T) {
		assert.False(t, (&Signer{}).Valid(expires, params, token))
	})
}
