package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/header"
)

func TestCreateOAuthToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("InvalidClientID", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"123"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("WrongClient", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2yy6"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("WrongSecret", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0f"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("AuthNotEnabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5gfsvbd7ejzn8m"},
			"client_secret": {"aaCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("UnknownAuthMethod", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2jh6"},
			"client_secret": {"aaCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"*"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestDeleteOAuthToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)
		DeleteOAuthToken(router)

		var tokenPath = "/api/v1/oauth/token"
		var logoutPath = "/api/v1/oauth/logout"

		data := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		createToken, _ := http.NewRequest("POST", tokenPath, strings.NewReader(data.Encode()))
		createToken.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		createResp := httptest.NewRecorder()
		app.ServeHTTP(createResp, createToken)

		t.Logf("Header: %s", createResp.Header())
		t.Logf("BODY: %s", createResp.Body.String())
		assert.Equal(t, http.StatusOK, createResp.Code)
		authToken := gjson.Get(createResp.Body.String(), "access_token").String()

		deleteToken, _ := http.NewRequest("POST", logoutPath, nil)
		deleteToken.Header.Add(header.XAuthToken, authToken)

		deleteResp := httptest.NewRecorder()
		app.ServeHTTP(deleteResp, deleteToken)

		t.Logf("Header: %s", deleteResp.Header())
		t.Logf("BODY: %s", deleteResp.Body.String())
		assert.Equal(t, http.StatusOK, deleteResp.Code)
	})
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DeleteOAuthToken(router)

		sess := entity.SessionFixtures.Get("alice_token")

		deleteToken, _ := http.NewRequest("POST", "/api/v1/oauth/logout", nil)
		deleteToken.Header.Add(header.XAuthToken, sess.AuthToken())

		deleteResp := httptest.NewRecorder()
		app.ServeHTTP(deleteResp, deleteToken)

		t.Logf("Header: %s", deleteResp.Header())
		t.Logf("BODY: %s", deleteResp.Body.String())
		assert.Equal(t, http.StatusForbidden, deleteResp.Code)
	})
}
