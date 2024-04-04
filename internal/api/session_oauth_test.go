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
	"github.com/photoprism/photoprism/internal/form"
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
		req.Header.Add(header.ContentType, header.ContentTypeForm)

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
		req.Header.Add(header.ContentType, header.ContentTypeForm)

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
		req.Header.Add(header.ContentType, header.ContentTypeForm)

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
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("InvalidSecret", func(t *testing.T) {
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
		req.Header.Add(header.ContentType, header.ContentTypeForm)

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
		req.Header.Add(header.ContentType, header.ContentTypeForm)

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
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRevokeOAuthToken(t *testing.T) {
	const tokenPath = "/api/v1/oauth/token"
	const revokePath = "/api/v1/oauth/revoke"

	t.Run("Success", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)
		RevokeOAuthToken(router)

		data := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		createToken, _ := http.NewRequest("POST", tokenPath, strings.NewReader(data.Encode()))
		createToken.Header.Add(header.ContentType, header.ContentTypeForm)

		createResp := httptest.NewRecorder()
		app.ServeHTTP(createResp, createToken)

		t.Logf("Header: %s", createResp.Header())
		t.Logf("BODY: %s", createResp.Body.String())
		assert.Equal(t, http.StatusOK, createResp.Code)
		authToken := gjson.Get(createResp.Body.String(), "access_token").String()

		revokeToken, _ := http.NewRequest("POST", revokePath, nil)
		revokeToken.Header.Add(header.XAuthToken, authToken)

		revokeResp := httptest.NewRecorder()
		app.ServeHTTP(revokeResp, revokeToken)

		t.Logf("Header: %s", revokeResp.Header())
		t.Logf("BODY: %s", revokeResp.Body.String())
		assert.Equal(t, http.StatusOK, revokeResp.Code)
	})
	t.Run("FormData", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)
		RevokeOAuthToken(router)

		createData := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		createToken, _ := http.NewRequest("POST", tokenPath, strings.NewReader(createData.Encode()))
		createToken.Header.Add(header.ContentType, header.ContentTypeForm)

		createResp := httptest.NewRecorder()
		app.ServeHTTP(createResp, createToken)

		t.Logf("Header: %s", createResp.Header())
		t.Logf("BODY: %s", createResp.Body.String())
		assert.Equal(t, http.StatusOK, createResp.Code)
		authToken := gjson.Get(createResp.Body.String(), "access_token").String()

		revokeData := url.Values{
			"token":           {authToken},
			"token_type_hint": {form.ClientAccessToken},
		}

		revokeToken, _ := http.NewRequest("POST", revokePath, strings.NewReader(revokeData.Encode()))
		revokeToken.Header.Add(header.ContentType, header.ContentTypeForm)

		revokeResp := httptest.NewRecorder()
		app.ServeHTTP(revokeResp, revokeToken)

		t.Logf("Header: %s", revokeResp.Header())
		t.Logf("BODY: %s", revokeResp.Body.String())
		assert.Equal(t, http.StatusOK, revokeResp.Code)
	})
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		RevokeOAuthToken(router)

		sess := entity.SessionFixtures.Get("alice_token")

		revokeToken, _ := http.NewRequest("POST", revokePath, nil)
		revokeToken.Header.Add(header.XAuthToken, sess.AuthToken())

		revokeResp := httptest.NewRecorder()
		app.ServeHTTP(revokeResp, revokeToken)

		t.Logf("Header: %s", revokeResp.Header())
		t.Logf("BODY: %s", revokeResp.Body.String())
		assert.Equal(t, http.StatusForbidden, revokeResp.Code)
	})
}
