package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func TestNonce(t *testing.T) {
	t.Run("Unique", func(t *testing.T) {
		a, err := Nonce()
		require.NoError(t, err)
		b, err := Nonce()
		require.NoError(t, err)
		assert.NotEmpty(t, a)
		assert.NotEmpty(t, b)
		assert.NotEqual(t, a, b)
	})
	t.Run("UrlSafe", func(t *testing.T) {
		n, err := Nonce()
		require.NoError(t, err)
		// base64url uses A-Z, a-z, 0-9, '-' and '_' only (no padding).
		assert.NotContains(t, n, "+")
		assert.NotContains(t, n, "/")
		assert.NotContains(t, n, "=")
	})
}

func tokensWithNonce(nonce string) *oidc.Tokens[*oidc.IDTokenClaims] {
	return &oidc.Tokens[*oidc.IDTokenClaims]{
		IDTokenClaims: &oidc.IDTokenClaims{
			TokenClaims: oidc.TokenClaims{Nonce: nonce},
		},
	}
}

func TestCheckNonce(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		assert.NoError(t, CheckNonce("abc", tokensWithNonce("abc")))
	})
	t.Run("Absent", func(t *testing.T) {
		// A provider that omits the nonce (e.g. a Cognito session-resumed token)
		// must not fail verification, even though we sent one.
		assert.NoError(t, CheckNonce("abc", tokensWithNonce("")))
	})
	t.Run("Mismatch", func(t *testing.T) {
		assert.ErrorIs(t, CheckNonce("abc", tokensWithNonce("xyz")), ErrNonceMismatch)
	})
	t.Run("UnexpectedNonce", func(t *testing.T) {
		// A nonce echoed without an expected value (e.g. a lost nonce cookie) is
		// unverifiable and must be rejected rather than silently accepted.
		assert.ErrorIs(t, CheckNonce("", tokensWithNonce("xyz")), ErrNonceMismatch)
	})
	t.Run("NilTokens", func(t *testing.T) {
		assert.NoError(t, CheckNonce("abc", nil))
	})
	t.Run("NilClaims", func(t *testing.T) {
		assert.NoError(t, CheckNonce("abc", &oidc.Tokens[*oidc.IDTokenClaims]{}))
	})
}
