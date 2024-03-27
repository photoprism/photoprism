package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
)

func TestEcho(t *testing.T) {
	t.Run("GET", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		Echo(router)

		authToken := AuthenticateAdmin(app, router)

		t.Logf("Auth Token: %s", authToken)
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/echo", authToken)
		t.Logf("Response Body: %s", r.Body.String())

		body := r.Body.String()
		url := gjson.Get(body, "url").String()
		method := gjson.Get(body, "method").String()
		request := gjson.Get(body, "headers.request")
		response := gjson.Get(body, "headers.response")

		assert.Equal(t, "/api/v1/echo", url)
		assert.Equal(t, "GET", method)
		assert.Equal(t, "Bearer "+authToken, request.Get("Authorization.0").String())
		assert.Equal(t, "application/json; charset=utf-8", response.Get("Content-Type.0").String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("POST", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		Echo(router)

		authToken := AuthenticateAdmin(app, router)

		t.Logf("Auth Token: %s", authToken)
		r := AuthenticatedRequest(app, http.MethodPost, "/api/v1/echo", authToken)

		body := r.Body.String()
		url := gjson.Get(body, "url").String()
		method := gjson.Get(body, "method").String()
		request := gjson.Get(body, "headers.request")
		response := gjson.Get(body, "headers.response")

		assert.Equal(t, "/api/v1/echo", url)
		assert.Equal(t, "POST", method)
		assert.Equal(t, "Bearer "+authToken, request.Get("Authorization.0").String())
		assert.Equal(t, "application/json; charset=utf-8", response.Get("Content-Type.0").String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
