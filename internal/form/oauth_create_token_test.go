package form

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/authn"
)

func TestOAuthCreateToken_Validate(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		m := OAuthCreateToken{
			ClientID:     "cs5gfen1bgxz7s9i",
			ClientSecret: "abc",
			Scope:        "*",
		}

		assert.NoError(t, m.Validate())
		assert.Equal(t, "*", m.CleanScope())
	})
	t.Run("GrantType", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:    authn.GrantClientCredentials,
			ClientID:     "cs5gfen1bgxz7s9i",
			ClientSecret: "abc",
			Scope:        "*",
		}

		assert.NoError(t, m.Validate())
		assert.Equal(t, "*", m.CleanScope())
	})
	t.Run("NoClientID", func(t *testing.T) {
		m := OAuthCreateToken{
			ClientID:     "",
			ClientSecret: "Alice123!",
			Scope:        "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("InvalidClientID", func(t *testing.T) {
		m := OAuthCreateToken{
			ClientID:     "s5gfen1bgxz7s9i",
			ClientSecret: "Alice123!",
			Scope:        "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("NoSecret", func(t *testing.T) {
		m := OAuthCreateToken{
			ClientID:     "cs5gfen1bgxz7s9i",
			ClientSecret: "",
			Scope:        "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("InvalidSecret", func(t *testing.T) {
		m := OAuthCreateToken{
			ClientID:     "cs5gfen1bgxz7s9i",
			ClientSecret: "abc  123",
			Scope:        "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("Password", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType: authn.GrantPassword,
			Username:  "admin",
			Password:  "cs5gfen1bgxz7s9i",
			Name:      "test",
			Scope:     "*",
		}

		assert.NoError(t, m.Validate())
	})
	t.Run("PasswordRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType: authn.GrantPassword,
			Username:  "admin",
			Password:  "",
			Name:      "test",
			Scope:     "*",
		}

		assert.Error(t, m.Validate())
	})
}
