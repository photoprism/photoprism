package oidc

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestNewClient(t *testing.T) {
	t.Run("Prod", func(t *testing.T) {
		uri, err := url.Parse("http://dummy-oidc:9998")

		assert.NoError(t, err)

		client, err := NewClient(
			uri,
			"csg6yqvykh0780f9",
			"nd09wkee0ElsMvzLGkgWS9wJAttHwF2h",
			authn.OidcDefaultScopes,
			"",
			"https://app.localssl.dev/",
			false,
		)

		assert.Error(t, err)
		assert.Nil(t, client)
	})
	t.Run("Debug", func(t *testing.T) {
		uri, err := url.Parse("http://dummy-oidc:9998")

		assert.NoError(t, err)

		client, err := NewClient(
			uri,
			"csg6yqvykh0780f9",
			"nd09wkee0ElsMvzLGkgWS9wJAttHwF2h",
			authn.OidcDefaultScopes,
			"",
			"https://app.localssl.dev/",
			true,
		)

		assert.NoError(t, err)
		assert.IsType(t, &Client{}, client)
	})
	t.Run("EmptyScopes", func(t *testing.T) {
		uri, err := url.Parse("http://dummy-oidc:9998")

		assert.NoError(t, err)

		client, err := NewClient(
			uri,
			"csg6yqvykh0780f9",
			"nd09wkee0ElsMvzLGkgWS9wJAttHwF2h",
			"",
			"",
			"https://app.localssl.dev/",
			true,
		)

		assert.NoError(t, err)
		assert.IsType(t, &Client{}, client)
	})
	t.Run("IssuerUriMissing", func(t *testing.T) {
		client, err := NewClient(
			nil,
			"csg6yqvykh0780f9",
			"nd09wkee0ElsMvzLGkgWS9wJAttHwF2h",
			authn.OidcDefaultScopes,
			"",
			"https://app.localssl.dev/",
			true,
		)

		assert.Error(t, err)
		assert.Nil(t, client)
	})
	t.Run("EmptyRedirectURL", func(t *testing.T) {
		uri, parseErr := url.Parse("http://dummy-oidc:9998")

		assert.NoError(t, parseErr)

		client, _ := NewClient(
			uri,
			"csg6yqvykh0780f9",
			"nd09wkee0ElsMvzLGkgWS9wJAttHwF2h",
			authn.OidcDefaultScopes,
			"",
			"",
			true,
		)

		assert.Nil(t, client)
	})
	t.Run("ServiceDiscoveryFails", func(t *testing.T) {
		uri, err := url.Parse("https://dummy-oidc:9998")

		assert.NoError(t, err)

		client, err := NewClient(
			uri,
			"csg6yqvykh0780f9",
			"nd09wkee0ElsMvzLGkgWS9wJAttHwF2h",
			authn.OidcDefaultScopes,
			"",
			"https://app.localssl.dev/",
			true,
		)

		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestNewClient_LogsInvalidPrompt(t *testing.T) {
	// Capture the audit log so the "unsupported prompt" warning can be asserted.
	var buf bytes.Buffer
	orig := event.AuditLog
	testLog := logrus.New()
	testLog.SetOutput(&buf)
	testLog.SetLevel(logrus.TraceLevel)
	event.AuditLog = testLog
	t.Cleanup(func() { event.AuditLog = orig })

	uri, err := url.Parse("http://dummy-oidc:9998")
	require.NoError(t, err)
	client, err := NewClient(uri, "csg6yqvykh0780f9", "nd09wkee0ElsMvzLGkgWS9wJAttHwF2h", authn.OidcDefaultScopes, "login bogus", "https://app.localssl.dev/", true)
	require.NoError(t, err)
	require.NotNil(t, client)

	// The unsupported token is reported; the valid one still reaches the provider.
	out := buf.String()
	assert.Contains(t, out, "prompt")
	assert.Contains(t, out, "bogus")
	assert.Equal(t, []string{"login"}, client.prompt)
}

func TestCodeExchangeRecorder(t *testing.T) {
	t.Run("CapturesStatusAndHeadersDiscardsBody", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		rec.Header().Set("oidc_error", "boom")
		rec.WriteHeader(http.StatusBadRequest)
		n, err := rec.Write([]byte("failed to get state: http: named cookie not present"))
		require.NoError(t, err)
		assert.Equal(t, len("failed to get state: http: named cookie not present"), n)
		assert.Equal(t, http.StatusBadRequest, rec.status)
		assert.Equal(t, "boom", rec.header.Get("oidc_error"))
	})
	t.Run("WriteWithoutWriteHeaderDefaultsTo200", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		_, _ = rec.Write([]byte("ok"))
		assert.Equal(t, http.StatusOK, rec.status)
	})
}

func TestClient_AuthURLHandler_SendsNonce(t *testing.T) {
	uri, err := url.Parse("http://dummy-oidc:9998")
	require.NoError(t, err)
	client, err := NewClient(uri, "csg6yqvykh0780f9", "nd09wkee0ElsMvzLGkgWS9wJAttHwF2h", authn.OidcDefaultScopes, "", "https://app.localssl.dev/", true)
	require.NoError(t, err)
	require.NotNil(t, client)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/oidc/login", nil)

	client.AuthURLHandler(c)

	// Redirects to the provider with a nonce on the authorization request.
	assert.Equal(t, http.StatusFound, w.Code)
	loc, locErr := url.Parse(w.Header().Get("Location"))
	require.NoError(t, locErr)
	sentNonce := loc.Query().Get("nonce")
	assert.NotEmpty(t, sentNonce)

	// Omits the prompt parameter when no authorization prompt is configured.
	assert.Empty(t, loc.Query().Get("prompt"))

	// Stores the nonce in a cookie so it survives to the callback.
	var nonceCookie bool
	for _, ck := range w.Result().Cookies() {
		if ck.Name == NonceCookie {
			nonceCookie = true
			assert.NotEmpty(t, ck.Value)
		}
	}
	assert.True(t, nonceCookie)
}

func TestClient_AuthURLHandler_SendsPrompt(t *testing.T) {
	t.Run("ValidValue", func(t *testing.T) {
		uri, err := url.Parse("http://dummy-oidc:9998")
		require.NoError(t, err)
		client, err := NewClient(uri, "csg6yqvykh0780f9", "nd09wkee0ElsMvzLGkgWS9wJAttHwF2h", authn.OidcDefaultScopes, "select_account", "https://app.localssl.dev/", true)
		require.NoError(t, err)
		require.NotNil(t, client)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/oidc/login", nil)

		client.AuthURLHandler(c)

		// Forwards the configured prompt on the authorization request.
		assert.Equal(t, http.StatusFound, w.Code)
		loc, locErr := url.Parse(w.Header().Get("Location"))
		require.NoError(t, locErr)
		assert.Equal(t, "select_account", loc.Query().Get("prompt"))
	})
	t.Run("InvalidValueIgnored", func(t *testing.T) {
		uri, err := url.Parse("http://dummy-oidc:9998")
		require.NoError(t, err)
		client, err := NewClient(uri, "csg6yqvykh0780f9", "nd09wkee0ElsMvzLGkgWS9wJAttHwF2h", authn.OidcDefaultScopes, "bogus", "https://app.localssl.dev/", true)
		require.NoError(t, err)
		require.NotNil(t, client)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/oidc/login", nil)

		client.AuthURLHandler(c)

		// An unsupported prompt value is dropped and never breaks the redirect.
		assert.Equal(t, http.StatusFound, w.Code)
		loc, locErr := url.Parse(w.Header().Get("Location"))
		require.NoError(t, locErr)
		assert.Empty(t, loc.Query().Get("prompt"))
	})
}

func TestClient_CodeExchangeUserInfo_NoStateCookie(t *testing.T) {
	// A redirect callback without the RP state cookie (e.g. an expired/interrupted
	// login) must return an error AND leave the real response untouched, so the
	// caller can render a branded page instead of the raw zitadel error.
	uri, err := url.Parse("http://dummy-oidc:9998")
	require.NoError(t, err)
	client, err := NewClient(uri, "csg6yqvykh0780f9", "nd09wkee0ElsMvzLGkgWS9wJAttHwF2h", authn.OidcDefaultScopes, "", "https://app.localssl.dev/", true)
	require.NoError(t, err)
	require.NotNil(t, client)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/oidc/redirect?code=abc&state=xyz", nil)

	userInfo, tokens, exErr := client.CodeExchangeUserInfo(c)
	assert.Error(t, exErr)
	assert.Nil(t, userInfo)
	assert.Nil(t, tokens)
	// The recorder absorbed the handler's raw error; the real writer is clean.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
}
