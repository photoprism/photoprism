package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestParseOAuthAuthorizeParams(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet,
		"/?response_type=code&client_id=cabc&redirect_uri=https%3A%2F%2Fapp%2Fcb&scope=openid%20profile&state=s1&code_challenge=ch&code_challenge_method=S256&nonce=n1", nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	p := parseOAuthAuthorizeParams(c)
	assert.Equal(t, "code", p.ResponseType)
	assert.Equal(t, "cabc", p.ClientID)
	assert.Equal(t, "https://app/cb", p.RedirectURI)
	assert.Equal(t, "openid profile", p.Scope)
	assert.Equal(t, "s1", p.State)
	assert.Equal(t, "ch", p.CodeChallenge)
	assert.Equal(t, "S256", p.CodeChallengeMethod)
	assert.Equal(t, "n1", p.Nonce)
}

func TestValidateOAuthAuthorizeRequest(t *testing.T) {
	valid := func() *oauthAuthorizeParams {
		return &oauthAuthorizeParams{
			ResponseType:        "code",
			ClientID:            "cabc",
			RedirectURI:         "https://app/cb",
			Scope:               "openid",
			State:               "s1",
			CodeChallenge:       "ch",
			CodeChallengeMethod: "S256",
		}
	}
	t.Run("Success", func(t *testing.T) {
		assert.NoError(t, validateOAuthAuthorizeRequest(valid()))
	})
	t.Run("MissingResponseType", func(t *testing.T) {
		p := valid()
		p.ResponseType = ""
		assert.Error(t, validateOAuthAuthorizeRequest(p))
	})
	t.Run("WrongResponseType", func(t *testing.T) {
		p := valid()
		p.ResponseType = "token"
		assert.Error(t, validateOAuthAuthorizeRequest(p))
	})
	t.Run("MissingScope", func(t *testing.T) {
		p := valid()
		p.Scope = ""
		assert.Error(t, validateOAuthAuthorizeRequest(p))
	})
	t.Run("MissingState", func(t *testing.T) {
		p := valid()
		p.State = ""
		assert.Error(t, validateOAuthAuthorizeRequest(p))
	})
	t.Run("MissingCodeChallenge", func(t *testing.T) {
		p := valid()
		p.CodeChallenge = ""
		assert.Error(t, validateOAuthAuthorizeRequest(p))
	})
	t.Run("MissingCodeChallengeMethod", func(t *testing.T) {
		p := valid()
		p.CodeChallengeMethod = ""
		assert.Error(t, validateOAuthAuthorizeRequest(p))
	})
	t.Run("WrongCodeChallengeMethod", func(t *testing.T) {
		p := valid()
		p.CodeChallengeMethod = "plain"
		assert.Error(t, validateOAuthAuthorizeRequest(p))
	})
}

func TestOAuthRedirectURIPermitted(t *testing.T) {
	client := &entity.Client{}
	client.SetData(&entity.ClientData{RedirectURIs: []string{"https://app.example.com/cb"}})
	t.Run("ExactMatch", func(t *testing.T) {
		assert.True(t, oauthRedirectURIPermitted(client, "https://app.example.com/cb"))
	})
	t.Run("Mismatch", func(t *testing.T) {
		assert.False(t, oauthRedirectURIPermitted(client, "https://app.example.com/other"))
	})
	t.Run("NilClient", func(t *testing.T) {
		assert.False(t, oauthRedirectURIPermitted(nil, "https://app.example.com/cb"))
	})
	t.Run("EmptyURI", func(t *testing.T) {
		assert.False(t, oauthRedirectURIPermitted(client, ""))
	})
	t.Run("NoRegisteredURIs", func(t *testing.T) {
		empty := &entity.Client{}
		empty.SetData(&entity.ClientData{})
		assert.False(t, oauthRedirectURIPermitted(empty, "https://app.example.com/cb"))
	})
}

func TestBuildOAuthRedirect(t *testing.T) {
	t.Run("AppendsParams", func(t *testing.T) {
		dest, err := buildOAuthRedirect("https://app.example.com/cb", map[string]string{"code": "abc", "state": "xyz"})
		require.NoError(t, err)
		u, err := url.Parse(dest)
		require.NoError(t, err)
		assert.Equal(t, "abc", u.Query().Get("code"))
		assert.Equal(t, "xyz", u.Query().Get("state"))
	})
	t.Run("PreservesExistingQuery", func(t *testing.T) {
		dest, err := buildOAuthRedirect("https://app.example.com/cb?foo=bar", map[string]string{"code": "abc"})
		require.NoError(t, err)
		u, err := url.Parse(dest)
		require.NoError(t, err)
		assert.Equal(t, "bar", u.Query().Get("foo"))
		assert.Equal(t, "abc", u.Query().Get("code"))
	})
	t.Run("DropsEmptyValues", func(t *testing.T) {
		dest, err := buildOAuthRedirect("https://app.example.com/cb", map[string]string{"code": "abc", "state": ""})
		require.NoError(t, err)
		u, err := url.Parse(dest)
		require.NoError(t, err)
		assert.Equal(t, "abc", u.Query().Get("code"))
		_, hasState := u.Query()["state"]
		assert.False(t, hasState, "empty value must be dropped")
	})
	t.Run("InvalidURI", func(t *testing.T) {
		_, err := buildOAuthRedirect("://bad", map[string]string{"code": "abc"})
		assert.Error(t, err)
	})
}

func TestOAuthScopeList(t *testing.T) {
	assert.Equal(t, []string{"openid", "profile"}, oauthScopeList("openid profile"))
	assert.Empty(t, oauthScopeList(""))
	assert.Empty(t, oauthScopeList("   "))
}

func TestOAuthClientDisplayName(t *testing.T) {
	t.Run("Named", func(t *testing.T) {
		c := &entity.Client{ClientName: "TestApp"}
		assert.Equal(t, c.Name(), oauthClientDisplayName(c))
		assert.NotEmpty(t, oauthClientDisplayName(c))
	})
	t.Run("FallsBackToUID", func(t *testing.T) {
		c := &entity.Client{ClientUID: "cs5gfen1bgxz7s9i"}
		assert.Equal(t, "cs5gfen1bgxz7s9i", oauthClientDisplayName(c))
	})
}

func TestOAuthResolveAuthorizeClient(t *testing.T) {
	client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
	t.Run("Success", func(t *testing.T) {
		got, code, desc := oauthResolveAuthorizeClient(&oauthAuthorizeParams{ClientID: client.GetUID(), RedirectURI: testRedirectURI})
		require.NotNil(t, got)
		assert.Empty(t, code)
		assert.Empty(t, desc)
	})
	t.Run("MissingClientID", func(t *testing.T) {
		got, code, _ := oauthResolveAuthorizeClient(&oauthAuthorizeParams{RedirectURI: testRedirectURI})
		assert.Nil(t, got)
		assert.Equal(t, oauthErrInvalidRequest, code)
	})
	t.Run("UnknownClient", func(t *testing.T) {
		got, code, _ := oauthResolveAuthorizeClient(&oauthAuthorizeParams{ClientID: "cs0000000000000z", RedirectURI: testRedirectURI})
		assert.Nil(t, got)
		assert.Equal(t, oauthErrInvalidClient, code)
	})
	t.Run("RedirectNotRegistered", func(t *testing.T) {
		got, code, _ := oauthResolveAuthorizeClient(&oauthAuthorizeParams{ClientID: client.GetUID(), RedirectURI: "https://evil.example.com/cb"})
		assert.Nil(t, got)
		assert.Equal(t, oauthErrInvalidRequest, code)
	})
}
