package form

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestOAuthRevokeToken_Empty(t *testing.T) {
	t.Run("AuthTokenAndTypeHintEmpty", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "",
			TokenTypeHint: "",
		}
		assert.True(t, m.Empty())
	})
	t.Run("AuthTokenNotEmpty", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "abc",
			TokenTypeHint: "",
		}
		assert.False(t, m.Empty())
	})
	t.Run("TypeHintNotEmpty", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "",
			TokenTypeHint: "test",
		}
		assert.False(t, m.Empty())
	})
}

//nolint:gosec // G101: Validation tests intentionally use inline token fixtures.
func TestOAuthRevokeToken_Validate(t *testing.T) {
	t.Run("AuthTokenEmpty", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "",
			TokenTypeHint: "test",
		}
		assert.Error(t, m.Validate())
	})
	t.Run("AuthTokenInvalid", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "abc   234",
			TokenTypeHint: "test",
		}
		assert.Error(t, m.Validate())
	})
	t.Run("UnsupportedToken", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "abc234",
			TokenTypeHint: "test",
		}
		assert.Error(t, m.Validate())
	})
	t.Run("AccessToken", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         rnd.AuthToken(),
			TokenTypeHint: "access_token",
		}
		assert.NoError(t, m.Validate())
		assert.Equal(t, AccessToken, m.TokenTypeHint)
	})
	t.Run("AccessTokenInvalid", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "abc",
			TokenTypeHint: "access_token",
		}
		assert.Error(t, m.Validate())
	})
	t.Run("SessionID", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         rnd.SessionID(rnd.AuthToken()),
			TokenTypeHint: "session_id",
		}
		assert.NoError(t, m.Validate())
		assert.Equal(t, SessionID, m.TokenTypeHint)
	})
	t.Run("SessionIDInvalid", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "abc",
			TokenTypeHint: "session_id",
		}
		assert.Error(t, m.Validate())
	})
	t.Run("RefID", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "sessxkkcabce",
			TokenTypeHint: "ref_id",
		}
		assert.NoError(t, m.Validate())
		assert.Equal(t, "ref_id", m.TokenTypeHint)
	})
	t.Run("RefIDInvalid", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "abc",
			TokenTypeHint: "ref_id",
		}
		assert.Error(t, m.Validate())
	})
	t.Run("NoTokenTypeHint", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         rnd.AuthToken(),
			TokenTypeHint: "",
		}
		assert.NoError(t, m.Validate())
		assert.Equal(t, AccessToken, m.TokenTypeHint)
	})
	t.Run("TypeHintEmptyRefID", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "sessxkkcabce",
			TokenTypeHint: "",
		}
		assert.NoError(t, m.Validate())
		assert.Equal(t, "ref_id", m.TokenTypeHint)
	})
	t.Run("TypeHintEmptySessionID", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         rnd.SessionID(rnd.AuthToken()),
			TokenTypeHint: "",
		}
		assert.NoError(t, m.Validate())
		assert.Equal(t, SessionID, m.TokenTypeHint)
	})
	t.Run("TypeHintEmptyAccessToken", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         rnd.AuthToken(),
			TokenTypeHint: "",
		}
		assert.NoError(t, m.Validate())
		assert.Equal(t, AccessToken, m.TokenTypeHint)
	})
	t.Run("MismatchedHint", func(t *testing.T) {
		m := OAuthRevokeToken{Token: rnd.SessionID(rnd.AuthToken()), TokenTypeHint: AccessToken}
		assert.NoError(t, m.Validate())
		assert.Equal(t, SessionID, m.TokenTypeHint)
		m = OAuthRevokeToken{Token: "sessxkkcabce", TokenTypeHint: SessionID}
		assert.NoError(t, m.Validate())
		assert.Equal(t, RefID, m.TokenTypeHint)
		m = OAuthRevokeToken{Token: rnd.AuthToken(), TokenTypeHint: RefID}
		assert.NoError(t, m.Validate())
		assert.Equal(t, AccessToken, m.TokenTypeHint)
	})
	t.Run("TypeHintEmptyInvalidToken", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "123",
			TokenTypeHint: "",
		}
		assert.Error(t, m.Validate())
	})
}

// TestOAuthRevokeToken_JSON pins every field to the JSON name the endpoint
// documents.
func TestOAuthRevokeToken_JSON(t *testing.T) {
	var m OAuthRevokeToken

	if err := json.Unmarshal([]byte(`{"token":"sess","token_type_hint":"session_id"}`), &m); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "sess", m.Token)
	assert.Equal(t, SessionID, m.TokenTypeHint)
	assert.False(t, m.Empty())
}

func TestOAuthRevokeToken_RefreshToken(t *testing.T) {
	token := rnd.RefreshToken(rnd.GenerateUID('g'))

	t.Run("Detected", func(t *testing.T) {
		f := OAuthRevokeToken{Token: token}
		assert.NoError(t, f.Validate())
		assert.Equal(t, RefreshToken, f.TokenTypeHint)
	})
	t.Run("Hint", func(t *testing.T) {
		f := OAuthRevokeToken{Token: token, TokenTypeHint: RefreshToken}
		assert.NoError(t, f.Validate())
	})
	t.Run("OtherHint", func(t *testing.T) {
		f := OAuthRevokeToken{Token: token, TokenTypeHint: AccessToken}
		assert.NoError(t, f.Validate())
		assert.Equal(t, RefreshToken, f.TokenTypeHint)
	})
	t.Run("AccessTokenWithHint", func(t *testing.T) {
		f := OAuthRevokeToken{Token: rnd.AuthToken(), TokenTypeHint: RefreshToken}
		assert.NoError(t, f.Validate())
		assert.Equal(t, AccessToken, f.TokenTypeHint)
	})
	t.Run("Signed", func(t *testing.T) {
		f := OAuthRevokeToken{Token: token + ".1700000000." + strings.Repeat("a", 32)}
		assert.NoError(t, f.Validate())
		assert.Equal(t, RefreshToken, f.TokenTypeHint)
	})
}

func TestIsRefreshTokenShape(t *testing.T) {
	token := rnd.RefreshToken(rnd.GenerateUID('g'))

	t.Run("Success", func(t *testing.T) {
		assert.True(t, isRefreshTokenShape(token))
		assert.True(t, isRefreshTokenShape(token+".1700000000."+strings.Repeat("0", 32)))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, isRefreshTokenShape(""))
		assert.False(t, isRefreshTokenShape(rnd.AuthToken()))
		assert.False(t, isRefreshTokenShape(rnd.AuthToken()+".1700000000."+strings.Repeat("0", 32)))
		assert.False(t, isRefreshTokenShape(token+".1700000000."+strings.Repeat("0", 31)))
		assert.False(t, isRefreshTokenShape(token+"."+strings.Repeat("0", 32)))
		assert.False(t, isRefreshTokenShape(token+".01700000000."+strings.Repeat("0", 32)))
	})
}
