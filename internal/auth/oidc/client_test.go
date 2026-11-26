package oidc

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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
	t.Run("CapturesStatusHeadersAndBody", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		rec.Header().Set("oidc_error", "boom")
		rec.WriteHeader(http.StatusBadRequest)
		n, err := rec.Write([]byte("failed to get state: http: named cookie not present\n"))
		require.NoError(t, err)
		assert.Equal(t, len("failed to get state: http: named cookie not present\n"), n)
		assert.Equal(t, http.StatusBadRequest, rec.status)
		assert.Equal(t, "boom", rec.header.Get("oidc_error"))
		assert.Equal(t, "failed to get state: http: named cookie not present", rec.cause())
	})
	t.Run("WriteWithoutWriteHeaderDefaultsTo200", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		_, _ = rec.Write([]byte("ok"))
		assert.Equal(t, http.StatusOK, rec.status)
	})
	t.Run("BoundsBody", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		first := []byte("failed to exchange token: " + strings.Repeat("a", codeExchangeBodyLimit-40) + " ")
		second := bytes.Repeat([]byte("b"), 100)
		n, err := rec.Write(first)
		require.NoError(t, err)
		assert.Equal(t, len(first), n)
		n, err = rec.Write(second)
		require.NoError(t, err)
		assert.Equal(t, len(second), n)
		n, err = rec.Write([]byte("c"))
		require.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Len(t, rec.body, codeExchangeBodyLimit)
		// The truncated body ends at the last delimiter, so the cut-off word is dropped.
		assert.Equal(t, strings.TrimSpace(string(first)), rec.cause())
	})
	t.Run("TruncatedUri", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		_, _ = rec.Write([]byte("http status not ok: 502 Bad Gateway " + strings.Repeat("x", codeExchangeBodyLimit-50) + " https://user:secret@example.com/"))
		assert.Len(t, rec.body, codeExchangeBodyLimit)
		assert.NotContains(t, rec.cause(), "user")
		assert.True(t, strings.HasPrefix(rec.cause(), "http status not ok: 502 Bad Gateway"))
	})
	t.Run("TruncatedWithoutDelimiter", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		_, _ = rec.Write(bytes.Repeat([]byte("a"), codeExchangeBodyLimit+10))
		assert.Equal(t, "", rec.cause())
	})
	t.Run("MasksRequestValues", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		secret := "nd09wkee0Els+MvzLG/kgWS9wJ"
		_, _ = rec.Write([]byte("oauth2: cannot fetch token: 400 Bad Request\nResponse: client_secret=" + url.QueryEscape(secret) + "&code=c0de-value&code_verifier=v3rifier-value raw " + secret))
		cause := rec.cause(secret, "c0de-value", "v3rifier-value", "Bad", "")
		assert.NotContains(t, cause, "nd09wkee")
		assert.NotContains(t, cause, "c0de-value")
		assert.NotContains(t, cause, "v3rifier-value")
		assert.Equal(t, "oauth2: cannot fetch token: 400 Bad Request\nResponse: client_secret=***&code=***&code_verifier=*** raw ***", cause)
	})
	t.Run("RedactsUris", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		_, _ = rec.Write([]byte(`failed to exchange token: oauth2: "invalid_client" "AADSTS7000215" "https://admin:pw@login.example.com/error?code=7000215"`))
		cause := rec.cause()
		assert.NotContains(t, cause, "admin:pw")
		assert.NotContains(t, cause, "code=7000215")
		assert.Contains(t, cause, "AADSTS7000215")
		assert.Contains(t, cause, "login.example.com/error?***")
	})
	t.Run("OmitsUnparsedResponse", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		rec.WriteHeader(http.StatusUnauthorized)
		_, _ = rec.Write([]byte(`userinfo failed: failed to unmarshal response: invalid character 'e' looking for beginning of value eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ4In0.sig` + "\n"))
		assert.Equal(t, "userinfo failed: failed to unmarshal response", rec.cause())
	})
	t.Run("ExchangeErrorWithCause", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		rec.WriteHeader(http.StatusUnauthorized)
		_, _ = rec.Write([]byte("failed to exchange token: oauth2: \"invalid_client\" \"AADSTS7000215: Invalid client secret provided.\\r\\nTrace ID: 1\"\n\x1b[31msecond line\n"))
		logMsg, err := rec.exchangeError()
		require.Error(t, err)
		assert.Equal(t, "failed to exchange token for user info", err.Error())
		assert.True(t, strings.HasPrefix(logMsg, "failed to exchange token for user info: failed to exchange token: oauth2: 'invalid_client' 'AADSTS7000215: Invalid client secret provided."))
		assert.NotContains(t, logMsg, "\n")
		assert.NotContains(t, logMsg, "\r")
		assert.NotContains(t, logMsg, "\x1b")
		assert.NotContains(t, logMsg, "\\")
	})
	t.Run("ExchangeErrorUnparsedResponse", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		rec.WriteHeader(http.StatusUnauthorized)
		_, _ = rec.Write([]byte(`userinfo failed: failed to unmarshal response: invalid character 'e' looking for beginning of value eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ4In0.sig` + "\n"))
		logMsg, err := rec.exchangeError()
		require.Error(t, err)
		assert.Equal(t, "failed to exchange token for user info", err.Error())
		assert.Equal(t, "failed to exchange token for user info: userinfo failed: failed to unmarshal response", logMsg)
		assert.NotContains(t, logMsg, "eyJ")
	})
	t.Run("ExchangeErrorWithoutCause", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		rec.WriteHeader(http.StatusUnauthorized)
		logMsg, err := rec.exchangeError()
		require.Error(t, err)
		assert.Equal(t, "failed to exchange token for user info", err.Error())
		assert.Equal(t, "failed to exchange token for user info", logMsg)
	})
	t.Run("ExchangeErrorFromHeader", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		rec.WriteHeader(http.StatusInternalServerError)
		rec.Header().Add("oidc_error", "oidc: access denied")
		_, _ = rec.Write([]byte("ignored body"))
		logMsg, err := rec.exchangeError()
		require.Error(t, err)
		assert.Equal(t, "oidc: access denied", err.Error())
		assert.Equal(t, "oidc: access denied", logMsg)
	})
	t.Run("EmptyBody", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		rec.WriteHeader(http.StatusUnauthorized)
		assert.Equal(t, "", rec.cause())
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
	var buf bytes.Buffer
	orig := event.SystemLog
	testLog := logrus.New()
	testLog.SetOutput(&buf)
	testLog.SetLevel(logrus.TraceLevel)
	event.SystemLog = testLog
	t.Cleanup(func() { event.SystemLog = orig })

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
	require.Error(t, exErr)
	assert.Equal(t, "failed to exchange token for user info", exErr.Error())
	assert.Nil(t, userInfo)
	assert.Nil(t, tokens)
	// The recorder absorbed the handler's raw error; the real writer is clean.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
	// The log names the cause recorded from the handler's response.
	out := buf.String()
	assert.Contains(t, out, "status 401")
	assert.Contains(t, out, "failed to exchange token for user info: failed to get state: http: named cookie not present")
	assert.Equal(t, 1, strings.Count(out, "\n"))
}

func TestClient_CodeExchangeUserInfo_MasksRequestValues(t *testing.T) {
	// A provider whose error page repeats the token request must not get the request values into the log.
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issuer":"` + srv.URL + `","authorization_endpoint":"` + srv.URL + `/authorize","token_endpoint":"` + srv.URL + `/token","userinfo_endpoint":"` + srv.URL + `/userinfo","jwks_uri":"` + srv.URL + `/keys","code_challenge_methods_supported":["S256"]}`))
		case "/token":
			body, _ := io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("<p>bad request: " + string(body) + "</p>"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	var buf bytes.Buffer
	orig := event.SystemLog
	testLog := logrus.New()
	testLog.SetOutput(&buf)
	testLog.SetLevel(logrus.TraceLevel)
	event.SystemLog = testLog
	t.Cleanup(func() { event.SystemLog = orig })

	uri, err := url.Parse(srv.URL)
	require.NoError(t, err)
	secret := "nd09wkee0ElsMvzLGkgWS9wJAttHwF2h"
	client, err := NewClient(uri, "csg6yqvykh0780f9", secret, authn.OidcDefaultScopes, "", "https://app.localssl.dev/", true)
	require.NoError(t, err)
	require.True(t, client.IsPKCE())

	gin.SetMode(gin.TestMode)
	login := httptest.NewRecorder()
	lc, _ := gin.CreateTestContext(login)
	lc.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/oidc/login", nil)
	client.AuthURLHandler(lc)
	authUrl, err := url.Parse(login.Header().Get("Location"))
	require.NoError(t, err)
	state := authUrl.Query().Get("state")
	require.NotEmpty(t, state)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/oidc/redirect?code=c0de-from-request&state="+url.QueryEscape(state), nil)

	for _, ck := range login.Result().Cookies() {
		c.Request.AddCookie(ck)
	}

	verifier, err := client.CookieHandler().CheckCookie(c.Request, pkceCookie)
	require.NoError(t, err)
	require.NotEmpty(t, verifier)

	_, _, exErr := client.CodeExchangeUserInfo(c)
	require.Error(t, exErr)
	assert.Equal(t, "failed to exchange token for user info", exErr.Error())

	out := buf.String()
	assert.Contains(t, out, "failed to exchange token for user info: failed to exchange token: oauth2: cannot fetch token: 400 Bad Request")
	assert.Contains(t, out, "bad request:")
	assert.NotContains(t, out, secret)
	assert.NotContains(t, out, "c0de-from-request")
	assert.NotContains(t, out, verifier)
}

func TestCodeExchangeMask(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mask := codeExchangeMask("client id", "se+cret/value", "c0de-value", "v3rifier-value")
		basic := base64.StdEncoding.EncodeToString([]byte("client+id:se%2Bcret%2Fvalue"))
		assert.Equal(t, []string{basic, "se+cret/value", "v3rifier-value", "c0de-value"}, mask)

		rec := &codeExchangeRecorder{header: make(http.Header)}
		_, _ = rec.Write([]byte("<p>Authorization: Basic " + basic + "</p>"))
		assert.NotContains(t, rec.cause(mask...), basic)
	})
	t.Run("NoSecret", func(t *testing.T) {
		assert.Equal(t, []string{"", "", "c0de-value"}, codeExchangeMask("client", "", "c0de-value", ""))
	})
	t.Run("CodeMatchingBasicPrefix", func(t *testing.T) {
		// A code equal to the part of the Basic value that depends on the public client ID only.
		secret := "nd09wkee0ElsMvzLGkgWS9wJAttHwF2h"
		basic := base64.StdEncoding.EncodeToString([]byte("csg6yqvykh0780f9:" + secret))
		code := basic[:20]
		rec := &codeExchangeRecorder{header: make(http.Header)}
		_, _ = rec.Write([]byte("<p>Authorization: Basic " + basic + "</p>"))
		cause := rec.cause(codeExchangeMask("csg6yqvykh0780f9", secret, code, "")...)
		tail := basic[20:]

		for i := 0; i+8 <= len(tail); i++ {
			assert.NotContains(t, cause, tail[i:i+8])
		}
	})
	t.Run("SecretInUriPath", func(t *testing.T) {
		secret := "nd09wkee0Els+MvzLG/kgWS9wJ"
		rec := &codeExchangeRecorder{header: make(http.Header)}
		_, _ = rec.Write([]byte("http status not ok: 400 Bad Request https://idp.example.com/err/" + url.QueryEscape(secret) + " " + secret))
		cause := rec.cause(codeExchangeMask("client", secret, "", "")...)
		assert.NotContains(t, cause, "nd09wkee")
		assert.Contains(t, cause, "https://idp.example.com/err/***")
	})
	t.Run("CodeMatchingScheme", func(t *testing.T) {
		rec := &codeExchangeRecorder{header: make(http.Header)}
		_, _ = rec.Write([]byte("http status not ok: 502 Bad Gateway https://admin:pw@proxy.example.com/"))
		assert.NotContains(t, rec.cause(codeExchangeMask("client", "", "https", "")...), "admin:pw")
	})
}
