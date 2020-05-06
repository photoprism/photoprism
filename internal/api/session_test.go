package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
)

func TestCreateSession(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CreateSession(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/session", `{"email": "photoprism", "password": "photoprism"}`)
		val2 := gjson.Get(r.Body.String(), "user.Email")
		assert.Equal(t, "photoprism@localhost", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("bad request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CreateSession(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/session", `{"email": 123, "password": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid password", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CreateSession(router, conf)
		r := PerformRequestWithBody(app, "POST", "/api/v1/session", `{"email": "photoprism", "password": "xxx"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Invalid password", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestDeleteSession(t *testing.T) {
	app, router, conf := NewApiTest()
	CreateSession(router, conf)
	r := PerformRequestWithBody(app, "POST", "/api/v1/session", `{"email": "photoprism", "password": "photoprism"}`)
	token := gjson.Get(r.Body.String(), "token")

	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		DeleteSession(router, conf)
		r := PerformRequest(app, "DELETE", "/api/v1/session/"+token.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
