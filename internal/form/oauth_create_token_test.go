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
			GrantType:  authn.GrantPassword,
			Username:   "admin",
			Password:   "cs5gfen1bgxz7s9i",
			ClientName: "test",
			Scope:      "*",
		}

		assert.NoError(t, m.Validate())
	})
	t.Run("UsernameRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "",
			Password:   "abcdefg",
			ClientName: "test",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("UsernameTooLong", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "aaaaabbbbbccccdddddfffffrrrrrttttttyyyyssssssssssdddddllllloooooooooowerty",
			Password:   "abcdefg",
			ClientName: "test",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("PasswordRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "admin",
			Password:   "",
			ClientName: "test",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("PasswordTooLong", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "admin",
			Password:   "aaaaabbbbbccccdddddfffffrrrrrttttttyyyyssssssssssdddddllllloooooooooowerty",
			ClientName: "test",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("ClientNameRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "admin",
			Password:   "cs5gfen1bgxz7s9i",
			ClientName: "",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("ScopeRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "admin",
			Password:   "cs5gfen1bgxz7s9i",
			ClientName: "test",
			Scope:      "",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("InvalidGrantType", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  "invalid",
			Username:   "admin",
			Password:   "cs5gfen1bgxz7s9i",
			ClientName: "test",
			Scope:      "",
		}

		assert.Error(t, m.Validate())
	})
}
