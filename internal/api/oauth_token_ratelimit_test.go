package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"golang.org/x/time/rate"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestOAuthToken_RateLimit_ClientCredentials(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	OAuthToken(router)

	// Tighten rate limits
	oldLogin, oldAuth := limiter.Login, limiter.Auth
	defer func() { limiter.Login, limiter.Auth = oldLogin, oldAuth }()
	limiter.Login = limiter.NewLimit(rate.Every(24*time.Hour), 3) // burst 3
	limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), 3)

	// Invalid client secret repeatedly (from UnknownIP: no headers set)
	path := "/api/v1/oauth/token"
	for i := 0; i < 3; i++ {
		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"INVALID"},
			"scope":         {"metrics"},
		}
		req, _ := http.NewRequest(http.MethodPost, path, strings.NewReader(data.Encode()))
		req.Header.Set(header.ContentType, header.ContentTypeForm)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
	// Next call should be rate limited
	data := url.Values{
		"grant_type":    {authn.GrantClientCredentials.String()},
		"client_id":     {"cs5cpu17n6gj2qo5"},
		"client_secret": {"INVALID"},
		"scope":         {"metrics"},
	}
	req, _ := http.NewRequest(http.MethodPost, path, strings.NewReader(data.Encode()))
	req.Header.Set(header.ContentType, header.ContentTypeForm)
	req.Header.Set("X-Forwarded-For", "198.51.100.99")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestOAuthToken_ResponseFields_ClientSuccess(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	OAuthToken(router)

	data := url.Values{
		"grant_type":    {authn.GrantClientCredentials.String()},
		"client_id":     {"cs5cpu17n6gj2qo5"},
		"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
		"scope":         {"metrics"},
	}
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Set(header.ContentType, header.ContentTypeForm)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.NotEmpty(t, gjson.Get(body, "access_token").String())
	tokType := gjson.Get(body, "token_type").String()
	assert.True(t, strings.EqualFold(tokType, "bearer"))
	assert.GreaterOrEqual(t, gjson.Get(body, "expires_in").Int(), int64(0))
	assert.Equal(t, "metrics", gjson.Get(body, "scope").String())
}

func TestOAuthToken_ResponseFields_UserSuccess(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	sess := AuthenticateUser(app, router, "alice", "Alice123!")
	OAuthToken(router)

	data := url.Values{
		"grant_type":  {authn.GrantPassword.String()},
		"client_name": {"TestApp"},
		"username":    {"alice"},
		"password":    {"Alice123!"},
		"scope":       {"*"},
	}
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Set(header.ContentType, header.ContentTypeForm)
	req.Header.Set(header.XAuthToken, sess)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.NotEmpty(t, gjson.Get(body, "access_token").String())
	tokType := gjson.Get(body, "token_type").String()
	assert.True(t, strings.EqualFold(tokType, "bearer"))
	assert.GreaterOrEqual(t, gjson.Get(body, "expires_in").Int(), int64(0))
	assert.Equal(t, "*", gjson.Get(body, "scope").String())
}

func TestOAuthToken_BadRequestsAndErrors(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	OAuthToken(router)

	// Missing grant_type & creds -> invalid credentials
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Unknown grant type
	data := url.Values{
		"grant_type": {"unknown"},
	}
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Set(header.ContentType, header.ContentTypeForm)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Password grant with wrong password
	sess := AuthenticateUser(app, router, "alice", "Alice123!")
	data = url.Values{
		"grant_type":  {authn.GrantPassword.String()},
		"client_name": {"AppPasswordAlice"},
		"username":    {"alice"},
		"password":    {"WrongPassword!"},
		"scope":       {"*"},
	}
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Set(header.ContentType, header.ContentTypeForm)
	req.Header.Set(header.XAuthToken, sess)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
