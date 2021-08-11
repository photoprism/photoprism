package api

import (
	"encoding/json"
	"github.com/photoprism/photoprism/internal/form"
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestCreateSession(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "admin", "password": "photoprism"}`)
		val2 := gjson.Get(r.Body.String(), "data.user.UserName")
		assert.Equal(t, "admin", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("bad request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": 123, "password": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("invalid token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "admin", "password": "photoprism", "token": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("valid token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "admin", "password": "photoprism", "token": "1jxf3jfn2k"}`)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("invalid password", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "admin", "password": "xxx"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrInvalidCredentials), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("alice - successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "alice", "password": "Alice123!"}`)
		resEmail := gjson.Get(r.Body.String(), "data.user.PrimaryEmail")
		resUsername := gjson.Get(r.Body.String(), "data.user.UserName")
		assert.Equal(t, "alice@example.com", resEmail.String())
		assert.Equal(t, "alice", resUsername.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("bob - successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "bob", "password": "Bobbob123!"}`)
		resEmail := gjson.Get(r.Body.String(), "data.user.PrimaryEmail")
		resUsername := gjson.Get(r.Body.String(), "data.user.UserName")
		assert.Equal(t, "bob@example.com", resEmail.String())
		assert.Equal(t, "bob", resUsername.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("bob - invalid password", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "bob", "password": "helloworld"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrInvalidCredentials), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestDeleteSession(t *testing.T) {
	t.Run("delete admin session", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		DeleteSession(router)
		f := form.Login{
			UserName: "admin",
			Password: "photoprism",
		}
		loginStr, err := json.Marshal(f)
		if err != nil {
			log.Fatal(err)
		}
		r0 := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", string(loginStr))
		sessId := r0.Header().Get("X-Session-ID")

		r := PerformRequest(app, http.MethodDelete, "/api/v1/session/"+sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("delete user session", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		DeleteSession(router)
		f := form.Login{
			UserName: "alice",
			Password: "Alice123!",
		}
		loginStr, err := json.Marshal(f)
		if err != nil {
			log.Fatal(err)
		}
		r0 := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", string(loginStr))
		sessId := r0.Header().Get("X-Session-ID")

		r := PerformRequest(app, http.MethodDelete, "/api/v1/session/"+sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("delete invalid session", func(t *testing.T) {
		sessId := "638bffc9b86a8fda0d908ebee84a43930cb8d1e3507f4aa0"
		app, router, _ := NewApiTest()
		DeleteSession(router)
		r := PerformRequest(app, http.MethodDelete, "/api/v1/session/"+sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
