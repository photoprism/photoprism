package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateOauthToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateOauthToken(router)

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

	t.Run("InvalidClientID", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateOauthToken(router)

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
		app, router, _ := NewApiTest()
		CreateOauthToken(router)

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
		app, router, _ := NewApiTest()
		CreateOauthToken(router)

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
		app, router, _ := NewApiTest()
		CreateOauthToken(router)

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
		app, router, _ := NewApiTest()
		CreateOauthToken(router)

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
