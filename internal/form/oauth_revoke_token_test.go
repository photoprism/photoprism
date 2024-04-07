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
	t.Run("SessionID", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         rnd.SessionID(rnd.AuthToken()),
			TokenTypeHint: "session_id",
		}
		assert.NoError(t, m.Validate())
		assert.Equal(t, SessionID, m.TokenTypeHint)
	})
	t.Run("NoTokenTypeHint", func(t *testing.T) {
		m := OAuthRevokeToken{
			Token:         rnd.AuthToken(),
			TokenTypeHint: "",
		}
		assert.NoError(t, m.Validate())
		assert.Equal(t, AccessToken, m.TokenTypeHint)
	})
}
