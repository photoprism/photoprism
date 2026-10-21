package oauth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/entity"
)

// testRequestSpec returns a valid pending request spec.
func testRequestSpec() RequestSpec {
	return RequestSpec{
		ClientID:            "cs5gfen1bgxz7s9i",
		ClientName:          "Test App",
		RedirectURI:         "https://claude.ai/api/mcp/auth_callback",
		Scopes:              []string{"mcp", "photos"},
		Resource:            "https://photos.example.com/api/v1/mcp",
		State:               "xyz",
		CodeChallenge:       testChallenge(testVerifier),
		CodeChallengeMethod: PKCEMethodS256,
	}
}

func TestNewRequest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		raw, err := NewRequest(testRequestSpec())
		require.NoError(t, err)

		m, err := FindRequest(raw)
		require.NoError(t, err)
		assert.Equal(t, "mcp photos", m.Scope)
		assert.Equal(t, "xyz", m.State)
		assert.Empty(t, m.UserUID)
		assert.WithinDuration(t, time.Now().UTC().Add(RequestTTL), m.ExpiresAt, 5*time.Second)

		// A request ID is not an authorization code.
		_, err = RedeemCode(raw, m.ClientID, m.RedirectURI, testVerifier)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
	t.Run("Incomplete", func(t *testing.T) {
		spec := testRequestSpec()
		spec.Scopes = nil
		_, err := NewRequest(spec)
		assert.Error(t, err)
	})
	t.Run("PlainChallenge", func(t *testing.T) {
		spec := testRequestSpec()
		spec.CodeChallengeMethod = "plain"
		_, err := NewRequest(spec)
		assert.Error(t, err)
	})
}

func TestFindRequest(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		_, err := FindRequest("0123456789abcdef0123456789abcdef0123456789abcdef")
		assert.ErrorIs(t, err, ErrRequestNotFound)
	})
	t.Run("Empty", func(t *testing.T) {
		_, err := FindRequest("")
		assert.ErrorIs(t, err, ErrRequestNotFound)
	})
	t.Run("Expired", func(t *testing.T) {
		raw, err := NewRequest(testRequestSpec())
		require.NoError(t, err)
		require.NoError(t, entity.UnscopedDb().Model(&entity.AuthCode{}).Where("code_hash = ?", HashSecret(raw)).
			UpdateColumn("expires_at", time.Now().UTC().Add(-time.Second)).Error)
		_, err = FindRequest(raw)
		assert.ErrorIs(t, err, ErrRequestNotFound)
	})
}

func TestRequestScopes(t *testing.T) {
	assert.Equal(t, []string{"mcp", "photos"}, RequestScopes(&entity.AuthCode{Scope: "mcp photos"}))
	assert.Nil(t, RequestScopes(nil))
}

func TestCompleteRequest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		raw, err := NewRequest(testRequestSpec())
		require.NoError(t, err)
		m, err := FindRequest(raw)
		require.NoError(t, err)

		code, err := CompleteRequest(m, "uqxetse3cy5eo9z2", []string{"mcp"})
		require.NoError(t, err)
		assert.NoError(t, tokens.VerifyCode(code))

		// The signed expiry is the stored one.
		stored, err := entity.FindAuthCode(HashSecret(code), entity.AuthCodeCode)
		require.NoError(t, err)
		require.NotNil(t, stored)
		_, expires, _, ok := tokens.SplitValue(code)
		require.True(t, ok)
		assert.Equal(t, stored.ExpiresAt.Unix(), expires)
		assert.Zero(t, stored.ExpiresAt.Nanosecond())

		// The request can no longer be found or completed.
		_, err = FindRequest(raw)
		assert.ErrorIs(t, err, ErrRequestNotFound)
		_, err = CompleteRequest(m, "uqxetse3cy5eo9z2", []string{"mcp"})
		assert.Error(t, err)

		redeemed, err := RedeemCode(code, m.ClientID, m.RedirectURI, testVerifier)
		require.NoError(t, err)
		assert.Equal(t, "uqxetse3cy5eo9z2", redeemed.UserUID)
		assert.Equal(t, "mcp", redeemed.Scope)
		assert.WithinDuration(t, time.Now().UTC().Add(CodeTTL), redeemed.ExpiresAt, 5*time.Second)
	})
	t.Run("ScopeNotRequested", func(t *testing.T) {
		raw, err := NewRequest(testRequestSpec())
		require.NoError(t, err)
		m, err := FindRequest(raw)
		require.NoError(t, err)
		_, err = CompleteRequest(m, "uqxetse3cy5eo9z2", []string{"mcp", "albums"})
		assert.Error(t, err)
		_, err = CompleteRequest(m, "uqxetse3cy5eo9z2", nil)
		assert.Error(t, err)
		_, err = CompleteRequest(m, "", []string{"mcp"})
		assert.Error(t, err)
	})
	t.Run("Expired", func(t *testing.T) {
		raw, err := NewRequest(testRequestSpec())
		require.NoError(t, err)
		m, err := FindRequest(raw)
		require.NoError(t, err)
		m.ExpiresAt = time.Now().UTC().Add(-time.Second)
		_, err = CompleteRequest(m, "uqxetse3cy5eo9z2", []string{"mcp"})
		assert.ErrorIs(t, err, ErrRequestNotFound)
	})
	t.Run("Nil", func(t *testing.T) {
		_, err := CompleteRequest(nil, "uqxetse3cy5eo9z2", []string{"mcp"})
		assert.ErrorIs(t, err, ErrRequestNotFound)
	})
}

func TestDenyRequest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		raw, err := NewRequest(testRequestSpec())
		require.NoError(t, err)
		m, err := FindRequest(raw)
		require.NoError(t, err)
		require.NoError(t, DenyRequest(m))
		_, err = FindRequest(raw)
		assert.ErrorIs(t, err, ErrRequestNotFound)
		assert.ErrorIs(t, DenyRequest(m), ErrRequestNotFound)
	})
	t.Run("Nil", func(t *testing.T) {
		assert.ErrorIs(t, DenyRequest(nil), ErrRequestNotFound)
	})
}
