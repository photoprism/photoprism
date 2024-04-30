package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/stretchr/testify/assert"
)

func TestCreateOAuthToken(t *testing.T) {
	t.Run("ClientSuccess", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
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
			"grant_type":    {authn.GrantClientCredentials.String()},
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
			"grant_type":    {authn.GrantClientCredentials.String()},
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
	t.Run("NotExistingClient", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"
		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
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
	t.Run("InvalidClientSecret", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
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
	t.Run("ClientAuthNotEnabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
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
	t.Run("ClientUnknownAuthMethod", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {"cs7pvt5h8rw9he34"},
			"client_secret": {"1831986451da7acf34690b703ff528f67bcf255e005270e9"},
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
	t.Run("UserNoSession", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":  {authn.GrantPassword.String()},
			"client_name": {"AppPasswordAlice"},
			"username":    {"alice"},
			"password":    {"Alice123!"},
			"scope":       {"*"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("UserSuccess", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":  {authn.GrantPassword.String()},
			"client_name": {"AppPasswordAlice"},
			"username":    {"alice"},
			"password":    {"Alice123!"},
			"scope":       {"*"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)
		req.Header.Add(header.XAuthToken, sessId)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("UnregisteredUser", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":  {authn.GrantPassword.String()},
			"client_name": {"Visitor"},
			"username":    {"visitor"},
			"password":    {"69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3"},
			"scope":       {"*"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)
		req.Header.Add(header.XAuthToken, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("UsersDontMatch", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":  {authn.GrantPassword.String()},
			"client_name": {"AppPasswordBob"},
			"username":    {"bob"},
			"password":    {"Bobbob123!"},
			"scope":       {"*"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)
		req.Header.Add(header.XAuthToken, sessId)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("DeletedUser", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		sessId := AuthenticateUser(app, router, "deleted", "Deleted123!")

		CreateOAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":  {authn.GrantPassword.String()},
			"client_name": {"AppPasswordDeleted"},
			"username":    {"deleted"},
			"password":    {"Deleted123!"},
			"scope":       {"*"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)
		req.Header.Add(header.XAuthToken, sessId)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
