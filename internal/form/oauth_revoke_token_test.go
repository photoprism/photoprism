package form

import (
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
	t.Run("TypeHintEmptyInvalidToken", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         "123",
			TokenTypeHint: "",
		}
		assert.Error(t, m.Validate())
	})
}
