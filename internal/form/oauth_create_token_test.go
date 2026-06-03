package form

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/authn"
)

//nolint:gosec // G101: Validation tests intentionally use inline OAuth credential fixtures.
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
	t.Run("GrantTypePasswordSuccess", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "admin",
			Password:   "cs5gfen1bgxz7s9i",
			ClientName: "test",
			Scope:      "*",
		}

		assert.NoError(t, m.Validate())
	})
	t.Run("GrantTypePasswordUsernameRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "",
			Password:   "abcdefg",
			ClientName: "test",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypePasswordUsernameTooLong", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "aaaaabbbbbccccdddddfffffrrrrrttttttyyyyssssssssssdddddllllloooooooooowerty",
			Password:   "abcdefg",
			ClientName: "test",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypePasswordPasswordRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "admin",
			Password:   "",
			ClientName: "test",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypePasswordPasswordTooLong", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "admin",
			Password:   "aaaaabbbbbccccdddddfffffrrrrrttttttyyyyssssssssssdddddllllloooooooooowerty",
			ClientName: "test",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypePasswordClientRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "admin",
			Password:   "cs5gfen1bgxz7s9i",
			ClientName: "",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypePasswordScopeRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantPassword,
			Username:   "admin",
			Password:   "cs5gfen1bgxz7s9i",
			ClientName: "test",
			Scope:      "",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypeSessionSuccess", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantSession,
			Username:   "admin",
			ClientName: "test",
			Scope:      "*",
		}

		assert.NoError(t, m.Validate())
	})
	t.Run("GrantTypeSessionUsernameRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantSession,
			Username:   "",
			ClientName: "test",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypeSessionClientRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantSession,
			Username:   "admin",
			ClientName: "",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypeSessionUsernameTooLong", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantSession,
			Username:   "aaaaabbbbbccccdddddfffffrrrrrttttttyyyyssssssssssdddddllllloooooooooowert",
			ClientName: "test",
			Scope:      "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypeSessionScopeRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:  authn.GrantSession,
			Username:   "admin",
			ClientName: "test",
			Scope:      "",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypeAuthorizationCodeSuccess", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:    authn.GrantAuthorizationCode,
			Code:         "authcode1234567890",
			CodeVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			RedirectURI:  "https://app.example.com/callback",
			ClientID:     "cs5gfen1bgxz7s9i",
		}

		assert.NoError(t, m.Validate())
	})
	t.Run("GrantTypeAuthorizationCodeCodeRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:    authn.GrantAuthorizationCode,
			Code:         "",
			CodeVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			RedirectURI:  "https://app.example.com/callback",
			ClientID:     "cs5gfen1bgxz7s9i",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypeAuthorizationCodeVerifierRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:   authn.GrantAuthorizationCode,
			Code:        "authcode1234567890",
			RedirectURI: "https://app.example.com/callback",
			ClientID:    "cs5gfen1bgxz7s9i",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypeAuthorizationCodeRedirectRequired", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:    authn.GrantAuthorizationCode,
			Code:         "authcode1234567890",
			CodeVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			ClientID:     "cs5gfen1bgxz7s9i",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("GrantTypeAuthorizationCodeInvalidClientID", func(t *testing.T) {
		m := OAuthCreateToken{
			GrantType:    authn.GrantAuthorizationCode,
			Code:         "authcode1234567890",
			CodeVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			RedirectURI:  "https://app.example.com/callback",
			ClientID:     "not-a-client-uid",
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
