package api

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
)

// PKCE pair from RFC 7636 Appendix B, reused across the authorize/token tests.
const (
	testPKCEVerifier  = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	testPKCEChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	testRedirectURI   = "https://app.example.com/callback"
)

// newOAuthAuthorizeClient registers a public client with the given redirect
// URIs and returns it.
func newOAuthAuthorizeClient(t *testing.T, redirectURIs []string) *entity.Client {
	client := entity.NewClient()
	client.SetName("Test OAuth App")
	client.ClientType = authn.ClientPublic
	client.SetScope("openid profile")
	client.AuthEnabled = true
	client.AuthExpires = 3600
	client.SetData(&entity.ClientData{RedirectURIs: redirectURIs})
	require.NoError(t, client.Create())
	return client
}

// oauthAuthorizeURL builds an authorize request query string.
func oauthAuthorizeURL(params map[string]string) string {
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	return "/api/v1/oauth/authorize?" + q.Encode()
}

func validAuthorizeParams(clientUID string) map[string]string {
	return map[string]string{
		"response_type":         "code",
		"client_id":             clientUID,
		"redirect_uri":          testRedirectURI,
		"scope":                 "openid profile",
		"state":                 "state-xyz",
		"code_challenge":        testPKCEChallenge,
		"code_challenge_method": "S256",
	}
}

func TestOAuthAuthorize_Get(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, conf := NewApiTest()
		app.LoadHTMLFiles(conf.TemplateFiles()...)
		OAuthAuthorize(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?response_type=code")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("ConsentPage", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		app.LoadHTMLFiles(conf.TemplateFiles()...)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthAuthorize(router)

		r := PerformRequest(app, http.MethodGet, oauthAuthorizeURL(validAuthorizeParams(client.GetUID())))
		assert.Equal(t, http.StatusOK, r.Code, "body=%s", r.Body.String())
		assert.Contains(t, r.Body.String(), "Authorize Access")
		assert.Contains(t, r.Body.String(), "Test OAuth App")
	})
	t.Run("UnknownClient", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		app.LoadHTMLFiles(conf.TemplateFiles()...)
		OAuthAuthorize(router)

		params := validAuthorizeParams("cs0000000000000z")
		r := PerformRequest(app, http.MethodGet, oauthAuthorizeURL(params))
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Contains(t, r.Body.String(), "Authorization Error")
	})
	t.Run("RedirectURINotRegistered", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		app.LoadHTMLFiles(conf.TemplateFiles()...)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthAuthorize(router)

		params := validAuthorizeParams(client.GetUID())
		params["redirect_uri"] = "https://evil.example.com/callback"
		r := PerformRequest(app, http.MethodGet, oauthAuthorizeURL(params))
		// MUST NOT redirect to an unregistered target.
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Empty(t, r.Header().Get("Location"))
	})
	t.Run("InvalidPKCEMethodRedirects", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		app.LoadHTMLFiles(conf.TemplateFiles()...)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthAuthorize(router)

		params := validAuthorizeParams(client.GetUID())
		params["code_challenge_method"] = "plain"
		r := PerformRequest(app, http.MethodGet, oauthAuthorizeURL(params))
		assert.Equal(t, http.StatusFound, r.Code)
		loc := r.Header().Get("Location")
		assert.Contains(t, loc, "error=invalid_request")
		assert.Contains(t, loc, "state=state-xyz")
	})
	t.Run("MissingState", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		app.LoadHTMLFiles(conf.TemplateFiles()...)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthAuthorize(router)

		params := validAuthorizeParams(client.GetUID())
		delete(params, "state")
		r := PerformRequest(app, http.MethodGet, oauthAuthorizeURL(params))
		assert.Equal(t, http.StatusFound, r.Code)
		assert.Contains(t, r.Header().Get("Location"), "error=invalid_request")
	})
}

func TestOAuthAuthorize_Post(t *testing.T) {
	t.Run("MintsCode", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthAuthorize(router)
		token := AuthenticateAdmin(app, router)

		r := AuthenticatedRequest(app, http.MethodPost, oauthAuthorizeURL(validAuthorizeParams(client.GetUID())), token)
		assert.Equal(t, http.StatusOK, r.Code, "body=%s", r.Body.String())
		redirect := gjson.Get(r.Body.String(), "redirect").String()
		require.NotEmpty(t, redirect)

		u, err := url.Parse(redirect)
		require.NoError(t, err)
		assert.Equal(t, "state-xyz", u.Query().Get("state"))
		code := u.Query().Get("code")
		require.NotEmpty(t, code)

		// The minted code must be bound to the client, user, and PKCE challenge.
		stored, err := entity.FindOAuthCode(code)
		require.NoError(t, err)
		require.NotNil(t, stored)
		assert.Equal(t, client.GetUID(), stored.ClientUID)
		assert.Equal(t, testRedirectURI, stored.RedirectURI)
		assert.Equal(t, testPKCEChallenge, stored.CodeChallenge)
		assert.NotEmpty(t, stored.UserUID)
		_ = stored.Delete()
	})
	t.Run("NoSession", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthAuthorize(router)

		r := PerformRequest(app, http.MethodPost, oauthAuthorizeURL(validAuthorizeParams(client.GetUID())))
		assert.Equal(t, http.StatusUnauthorized, r.Code)
		assert.Equal(t, "login_required", gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("RedirectURINotRegistered", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthAuthorize(router)
		token := AuthenticateAdmin(app, router)

		params := validAuthorizeParams(client.GetUID())
		params["redirect_uri"] = "https://evil.example.com/callback"
		r := AuthenticatedRequest(app, http.MethodPost, oauthAuthorizeURL(params), token)
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, "invalid_request", gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("InvalidParams", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthAuthorize(router)
		token := AuthenticateAdmin(app, router)

		params := validAuthorizeParams(client.GetUID())
		params["code_challenge_method"] = "plain"
		r := AuthenticatedRequest(app, http.MethodPost, oauthAuthorizeURL(params), token)
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, "invalid_request", gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()
		OAuthAuthorize(router)

		r := PerformRequest(app, http.MethodPost, "/api/v1/oauth/authorize")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
