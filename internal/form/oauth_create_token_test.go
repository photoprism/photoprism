package form

import (
	"encoding/json"
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

// TestOAuthCreateToken_JSON pins every field to the JSON name the endpoint
// documents.
//
//nolint:gosec // G101: Binding tests intentionally use inline OAuth credential fixtures.
func TestOAuthCreateToken_JSON(t *testing.T) {
	body := `{
		"grant_type": "client_credentials",
		"client_id": "cs5cpu17n6gj2qo5",
		"client_name": "test",
		"client_secret": "xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e",
		"username": "alice",
		"password": "photoprism",
		"refresh_token": "refresh",
		"code": "code",
		"code_verifier": "verifier",
		"redirect_uri": "https://photos.example.com/cb",
		"assertion": "assertion",
		"scope": "metrics",
		"expires_in": 3600
	}`

	var m OAuthCreateToken

	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, authn.GrantClientCredentials, m.GrantType)
	assert.Equal(t, "cs5cpu17n6gj2qo5", m.ClientID)
	assert.Equal(t, "test", m.ClientName)
	assert.Equal(t, "xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e", m.ClientSecret)
	assert.Equal(t, "alice", m.Username)
	assert.Equal(t, "photoprism", m.Password)
	assert.Equal(t, "refresh", m.RefreshToken)
	assert.Equal(t, "code", m.Code)
	assert.Equal(t, "verifier", m.CodeVerifier)
	assert.Equal(t, "https://photos.example.com/cb", m.RedirectURI)
	assert.Equal(t, "assertion", m.Assertion)
	assert.Equal(t, "metrics", m.Scope)
	assert.Equal(t, int64(3600), m.ExpiresIn)
}
