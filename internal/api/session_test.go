package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
)

func TestCreateSession(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/session", `{"username": "admin", "password": "photoprism"}`)
		val2 := gjson.Get(r.Body.String(), "user.Email")
		assert.Equal(t, "", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("bad request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/session", `{"username": 123, "password": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/session", `{"username": "admin", "password": "photoprism", "token": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("valid token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/session", `{"username": "admin", "password": "photoprism", "token": "1jxf3jfn2k"}`)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid password", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/session", `{"username": "admin", "password": "xxx"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Invalid user name or password", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestDeleteSession(t *testing.T) {
	app, router, _ := NewApiTest()
	CreateSession(router)
	r := PerformRequestWithBody(app, "POST", "/api/v1/session", `{"username": "admin", "password": "photoprism"}`)
	id := gjson.Get(r.Body.String(), "id")

	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		DeleteSession(router)
		r := PerformRequest(app, "DELETE", "/api/v1/session/"+id.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
