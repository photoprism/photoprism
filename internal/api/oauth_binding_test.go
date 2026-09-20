package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// bindOAuthTestRequest runs BindOAuthRequest against a request with the given
// content type and body, and returns the bound form.
func bindOAuthTestRequest(contentType, body string) (form.OAuthCreateToken, error) {
	return bindOAuthTestRequestURL("/api/v1/oauth/token", contentType, body)
}

// bindOAuthTestRequestURL is bindOAuthTestRequest with the request URL supplied,
// so a case can tell a value read from the body apart from one read from the query.
func bindOAuthTestRequestURL(url, contentType, body string) (form.OAuthCreateToken, error) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodPost, url, strings.NewReader(body))

	if contentType != "" {
		c.Request.Header.Set(header.ContentType, contentType)
	}

	var frm form.OAuthCreateToken
	err := BindOAuthRequest(c, &frm)

	return frm, err
}

// TestBindOAuthRequest covers which encodings an OAuth2 request body is decoded
// from: form, JSON and multipart populate the form, other content types do not.
func TestBindOAuthRequest(t *testing.T) {
	t.Run("Form", func(t *testing.T) {
		data := url.Values{
			"grant_type": {authn.GrantClientCredentials.String()},
			"client_id":  {"cs5cpu17n6gj2qo5"},
		}
		frm, err := bindOAuthTestRequest(header.ContentTypeForm, data.Encode())
		assert.NoError(t, err)
		assert.Equal(t, authn.GrantClientCredentials, frm.GrantType)
		assert.Equal(t, "cs5cpu17n6gj2qo5", frm.ClientID)
	})
	t.Run("Json", func(t *testing.T) {
		frm, err := bindOAuthTestRequest(header.ContentTypeJson, `{"grant_type":"client_credentials","client_id":"cs5cpu17n6gj2qo5"}`)
		assert.NoError(t, err)
		assert.Equal(t, authn.GrantClientCredentials, frm.GrantType)
		assert.Equal(t, "cs5cpu17n6gj2qo5", frm.ClientID)
	})
	t.Run("JsonWithCharset", func(t *testing.T) {
		frm, err := bindOAuthTestRequest(header.ContentTypeJsonUtf8, `{"grant_type":"client_credentials","client_id":"cs5cpu17n6gj2qo5"}`)
		assert.NoError(t, err)
		assert.Equal(t, "cs5cpu17n6gj2qo5", frm.ClientID)
	})
	t.Run("Multipart", func(t *testing.T) {
		body := "--x\r\nContent-Disposition: form-data; name=\"grant_type\"\r\n\r\nclient_credentials\r\n--x--\r\n"
		frm, err := bindOAuthTestRequestURL("/api/v1/oauth/token?grant_type=session", "multipart/form-data; boundary=x", body)
		assert.NoError(t, err)
		// The body carries the value, not the query string.
		assert.Equal(t, authn.GrantClientCredentials, frm.GrantType)
	})
	t.Run("Yaml", func(t *testing.T) {
		// Each body below is one the corresponding decoder does populate the
		// form from, so an empty form is evidence that the decoder did not run
		// rather than that the body failed to name the fields.
		body := "grant_type: client_credentials\nclient_id: cs5cpu17n6gj2qo5\nclient_secret: xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e\n"
		frm, err := bindOAuthTestRequest("application/yaml", body)
		assert.ErrorIs(t, err, ErrUnsupportedContentType)
		assert.Empty(t, frm.ClientID)
		assert.Empty(t, frm.ClientSecret)
		assert.Equal(t, authn.GrantUndefined, frm.GrantType)
	})
	t.Run("Toml", func(t *testing.T) {
		body := "ClientID = \"cs5cpu17n6gj2qo5\"\nGrantType = \"client_credentials\"\n"
		frm, err := bindOAuthTestRequest("application/toml", body)
		assert.ErrorIs(t, err, ErrUnsupportedContentType)
		assert.Empty(t, frm.ClientID)
		assert.Equal(t, authn.GrantUndefined, frm.GrantType)
	})
	t.Run("Xml", func(t *testing.T) {
		body := `<OAuthCreateToken><ClientID>cs5cpu17n6gj2qo5</ClientID></OAuthCreateToken>`
		frm, err := bindOAuthTestRequest("application/xml", body)
		assert.ErrorIs(t, err, ErrUnsupportedContentType)
		assert.Empty(t, frm.ClientID)
	})
	t.Run("JsonUppercase", func(t *testing.T) {
		// Media types are case-insensitive, so the decoder selection is too.
		frm, err := bindOAuthTestRequest("Application/JSON", `{"client_id":"cs5cpu17n6gj2qo5"}`)
		assert.NoError(t, err)
		assert.Equal(t, "cs5cpu17n6gj2qo5", frm.ClientID)
	})
	t.Run("QueryString", func(t *testing.T) {
		// Values come from the body. A query string is not a place to put a
		// credential, so one there does not authenticate.
		frm, err := bindOAuthTestRequestURL(
			"/api/v1/oauth/token?grant_type=client_credentials&client_id=cs5cpu17n6gj2qo5&client_secret=xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e",
			header.ContentTypeForm, "")
		assert.NoError(t, err)
		assert.Empty(t, frm.ClientID)
		assert.Empty(t, frm.ClientSecret)
		assert.Equal(t, authn.GrantUndefined, frm.GrantType)
	})
	t.Run("FormUppercase", func(t *testing.T) {
		// Media types are case-insensitive here as well.
		data := url.Values{"client_id": {"cs5cpu17n6gj2qo5"}}
		frm, err := bindOAuthTestRequest("Application/X-WWW-Form-Urlencoded", data.Encode())
		assert.NoError(t, err)
		assert.Equal(t, "cs5cpu17n6gj2qo5", frm.ClientID)
	})
	t.Run("Empty", func(t *testing.T) {
		// Without a content type the form decoder runs and leaves the body
		// unparsed: net/http reads a POST body as a form only for the
		// urlencoded and multipart types.
		frm, err := bindOAuthTestRequest("", "grant_type=client_credentials")
		assert.NoError(t, err)
		assert.Equal(t, authn.GrantUndefined, frm.GrantType)
	})
}

// TestOAuthToken_RequestEncoding covers the encodings the token endpoint accepts
// end to end, so the contract is pinned at the route and not only in the helper.
func TestOAuthToken_RequestEncoding(t *testing.T) {
	//nolint:gosec // G101: Static client fixture is intentional for token tests.
	const (
		clientID     = "cs5cpu17n6gj2qo5"
		clientSecret = "xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"
	)

	t.Run("Form", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {clientID},
			"client_secret": {clientSecret},
			"scope":         {"metrics"},
		}
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
		req.Header.Set(header.ContentType, header.ContentTypeForm)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("Yaml", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)

		// Complete credentials, named as the endpoint documents them, in an
		// encoding it does not accept. No token is issued for them.
		body := "grant_type: client_credentials\nclient_id: " + clientID + "\nclient_secret: " + clientSecret + "\nscope: metrics\n"
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(body))
		req.Header.Set(header.ContentType, "application/yaml")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.NotContains(t, w.Body.String(), "access_token")
	})
}

// TestOAuthRevoke_RequestEncoding covers a revoke request whose body is in an
// encoding the endpoint does not accept. The caller presents its own token in a
// header and names a different one in the body, so the response identifies which
// session a request in that shape reaches.
func TestOAuthRevoke_RequestEncoding(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	dropIsolatedClientSessions(t)

	OAuthToken(router)
	OAuthRevoke(router)

	// mint returns a fresh access token for the isolated test client.
	mint := func() string {
		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {isolatedClientID},
			"client_secret": {isolatedClientSecret},
			"scope":         {"metrics"},
		}
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
		req.Header.Set(header.ContentType, header.ContentTypeForm)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		return gjson.Get(w.Body.String(), "access_token").String()
	}

	caller, target := mint(), mint()
	assert.NotEmpty(t, caller)
	assert.NotEmpty(t, target)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/revoke",
		strings.NewReader("<r><Token>"+target+"</Token><TokenTypeHint>access_token</TokenTypeHint></r>"))
	req.Header.Set(header.ContentType, "application/xml")
	req.Header.Set(header.XAuthToken, caller)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// Neither session is revoked: the request is refused rather than applied to
	// the token the caller authenticated with.
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotEqual(t, rnd.SessionID(caller), gjson.Get(w.Body.String(), "session_id").String())

	callerSess, callerErr := entity.FindSessionByAuthToken(caller)
	assert.NoError(t, callerErr, "the caller session must survive")
	assert.NotNil(t, callerSess)

	targetSess, targetErr := entity.FindSessionByAuthToken(target)
	assert.NoError(t, targetErr, "the named session must survive")
	assert.NotNil(t, targetSess)
}

// TestOAuthToken_FormGateCasing covers a form request whose content type is cased
// differently. The grant gate and the binder read the encoding the same way, so a
// grant the Portal OIDC OP owns is still routed to it.
func TestOAuthToken_FormGateCasing(t *testing.T) {
	for _, contentType := range []string{
		header.ContentTypeForm,
		"Application/X-WWW-Form-Urlencoded",
		"APPLICATION/X-WWW-FORM-URLENCODED",
	} {
		t.Run(contentType, func(t *testing.T) {
			app, router, conf := NewApiTest()
			conf.SetAuthMode(config.AuthModePasswd)
			defer conf.SetAuthMode(config.AuthModePublic)

			prev := OAuthAuthorizationCodeHandler
			OAuthAuthorizationCodeHandler = nil
			defer func() { OAuthAuthorizationCodeHandler = prev }()

			OAuthToken(router)

			data := url.Values{
				"grant_type": {authn.GrantAuthorizationCode.String()},
				"code":       {"raw-code"},
			}
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
			req.Header.Set(header.ContentType, contentType)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			// The gate reports the grant rather than the binder reporting a credential.
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "unsupported_grant_type")
		})
	}
}

// TestOAuthToken_QueryCredentials covers credentials supplied in the query string
// rather than the request body.
func TestOAuthToken_QueryCredentials(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	OAuthToken(router)

	query := "?grant_type=client_credentials&client_id=cs5cpu17n6gj2qo5&client_secret=xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e&scope=metrics"
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token"+query, nil)
	req.Header.Set(header.ContentType, header.ContentTypeForm)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.NotContains(t, w.Body.String(), "access_token")
}
