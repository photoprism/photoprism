package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// postOAuthToken sends a form-encoded token request and returns the recorder.
func postOAuthToken(app *gin.Engine, data url.Values) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Add(header.ContentType, header.ContentTypeForm)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	return w
}

// mintTestAuthCode inserts a single-use authorization code and returns its raw
// value.
func mintTestAuthCode(t *testing.T, clientUID, userUID, redirectURI string) string {
	raw, _, err := entity.NewOAuthCode(entity.OAuthCodeSpec{
		ClientUID:           clientUID,
		UserUID:             userUID,
		RedirectURI:         redirectURI,
		Scope:               "openid profile",
		CodeChallenge:       testPKCEChallenge,
		CodeChallengeMethod: "S256",
	})
	require.NoError(t, err)
	return raw
}

func authCodeTokenForm(code, verifier, redirectURI, clientUID string) url.Values {
	return url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {verifier},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientUID},
	}
}

func TestOAuthToken_AuthorizationCode(t *testing.T) {
	admin := entity.FindUserByName("admin")
	require.NotNil(t, admin)

	t.Run("Success", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthToken(router)

		code := mintTestAuthCode(t, client.GetUID(), admin.UserUID, testRedirectURI)
		w := postOAuthToken(app, authCodeTokenForm(code, testPKCEVerifier, testRedirectURI, client.GetUID()))
		assert.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
		assert.NotEmpty(t, gjson.Get(w.Body.String(), "access_token").String())

		// Code must be single-use.
		gone, err := entity.FindOAuthCode(code)
		require.NoError(t, err)
		assert.Nil(t, gone)
	})
	t.Run("Replay", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthToken(router)

		code := mintTestAuthCode(t, client.GetUID(), admin.UserUID, testRedirectURI)
		first := postOAuthToken(app, authCodeTokenForm(code, testPKCEVerifier, testRedirectURI, client.GetUID()))
		require.Equal(t, http.StatusOK, first.Code, "body=%s", first.Body.String())

		replay := postOAuthToken(app, authCodeTokenForm(code, testPKCEVerifier, testRedirectURI, client.GetUID()))
		assert.Equal(t, http.StatusBadRequest, replay.Code)
		assert.Equal(t, "invalid_grant", gjson.Get(replay.Body.String(), "error").String())
	})
	t.Run("BadVerifier", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthToken(router)

		code := mintTestAuthCode(t, client.GetUID(), admin.UserUID, testRedirectURI)
		w := postOAuthToken(app, authCodeTokenForm(code, "wrong-verifier", testRedirectURI, client.GetUID()))
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid_grant", gjson.Get(w.Body.String(), "error").String())

		// A failed verifier still consumes the code (single attempt).
		gone, err := entity.FindOAuthCode(code)
		require.NoError(t, err)
		assert.Nil(t, gone)
	})
	t.Run("Expired", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthToken(router)

		code := mintTestAuthCode(t, client.GetUID(), admin.UserUID, testRedirectURI)
		stored, err := entity.FindOAuthCode(code)
		require.NoError(t, err)
		require.NotNil(t, stored)
		stored.ExpiresAt = time.Now().UTC().Add(-time.Hour)
		require.NoError(t, entity.Db().Save(stored).Error)

		w := postOAuthToken(app, authCodeTokenForm(code, testPKCEVerifier, testRedirectURI, client.GetUID()))
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid_grant", gjson.Get(w.Body.String(), "error").String())
	})
	t.Run("RedirectMismatch", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthToken(router)

		code := mintTestAuthCode(t, client.GetUID(), admin.UserUID, testRedirectURI)
		w := postOAuthToken(app, authCodeTokenForm(code, testPKCEVerifier, "https://app.example.com/other", client.GetUID()))
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid_grant", gjson.Get(w.Body.String(), "error").String())
	})
	t.Run("ClientMismatch", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		clientA := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		clientB := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthToken(router)

		code := mintTestAuthCode(t, clientA.GetUID(), admin.UserUID, testRedirectURI)
		w := postOAuthToken(app, authCodeTokenForm(code, testPKCEVerifier, testRedirectURI, clientB.GetUID()))
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid_grant", gjson.Get(w.Body.String(), "error").String())
	})
	t.Run("UnknownCode", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthToken(router)

		w := postOAuthToken(app, authCodeTokenForm("nonexistent-auth-code", testPKCEVerifier, testRedirectURI, client.GetUID()))
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid_grant", gjson.Get(w.Body.String(), "error").String())
	})
	t.Run("MissingVerifier", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI})
		OAuthToken(router)

		code := mintTestAuthCode(t, client.GetUID(), admin.UserUID, testRedirectURI)
		form := authCodeTokenForm(code, "", testRedirectURI, client.GetUID())
		w := postOAuthToken(app, form)
		// Form validation rejects the request before redemption (401).
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		_ = entity.Db().Where("code_hash = ?", entity.HashOAuthCode(code)).Delete(&entity.OAuthCode{})
	})
	t.Run("ExpiresInCapped", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		client := newOAuthAuthorizeClient(t, []string{testRedirectURI}) // AuthExpires = 3600
		OAuthToken(router)

		code := mintTestAuthCode(t, client.GetUID(), admin.UserUID, testRedirectURI)
		form := authCodeTokenForm(code, testPKCEVerifier, testRedirectURI, client.GetUID())
		form.Set("expires_in", "999999999")
		w := postOAuthToken(app, form)
		require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
		expiresIn := gjson.Get(w.Body.String(), "expires_in").Int()
		assert.Positive(t, expiresIn)
		assert.LessOrEqual(t, expiresIn, int64(3600), "token lifetime must be capped at the client's configured expiry")
	})
}
