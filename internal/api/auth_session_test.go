package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/session"
)

func TestSessionID(t *testing.T) {
	t.Run("NoContext", func(t *testing.T) {
		result := SessionID(nil)
		assert.Equal(t, "", result)
	})
}

func TestSession(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		assert.Equal(t, session.Public, Session(""))
		assert.Equal(t, session.Public, Session("638bffc9b86a8fda0d908ebee84a43930cb8d1e3507f4aa0"))
	})
}

func TestCreateSession(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "admin", "password": "photoprism"}`)
		log.Debugf("BODY: %s", r.Body.String())
		val2 := gjson.Get(r.Body.String(), "user.Name")
		assert.Equal(t, "admin", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("BadRequest", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": 123, "password": "xxx"}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("PublicInvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "admin", "password": "photoprism", "token": "xxx"}`)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("AdminInvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		// CreateSession(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"token": "xxx"}`, sessId)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("VisitorInvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateSession(router)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"token": "xxx"}`, "345346")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("AdminValidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"token": "1jxf3jfn2k"}`, sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("PublicValidToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "admin", "password": "photoprism", "token": "1jxf3jfn2k"}`)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("AdminInvalidPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateSession(router)

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", form.AsJson(form.Login{
			UserName: "admin",
			Password: "xxx",
		}))

		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrInvalidCredentials), val.String())
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("AliceSuccess", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateSession(router)

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "alice", "password": "Alice123!"}`)
		userEmail := gjson.Get(r.Body.String(), "user.Email")
		userName := gjson.Get(r.Body.String(), "user.Name")
		assert.Equal(t, "alice@example.com", userEmail.String())
		assert.Equal(t, "alice", userName.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("BobSuccess", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "bob", "password": "Bobbob123!"}`)
		userEmail := gjson.Get(r.Body.String(), "user.Email")
		userName := gjson.Get(r.Body.String(), "user.Name")
		assert.Equal(t, "bob@example.com", userEmail.String())
		assert.Equal(t, "bob", userName.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("BobInvalidPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateSession(router)
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "bob", "password": "helloworld"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrInvalidCredentials), val.String())
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
}

func TestGetSession(t *testing.T) {
	t.Run("AdminWithoutAuthentication", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetSession(router)

		sessId := AuthenticateAdmin(app, router)
		r := PerformRequest(app, http.MethodGet, "/api/v1/session/"+sessId)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("AdminAuthenticatedRequest", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetSession(router)

		sessId := AuthenticateAdmin(app, router)
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/session/"+sessId, sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestDeleteSession(t *testing.T) {
	t.Run("AdminWithoutAuthentication", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DeleteSession(router)

		sessId := AuthenticateAdmin(app, router)

		r := PerformRequest(app, http.MethodDelete, "/api/v1/session/"+sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("AdminAuthenticatedRequest", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DeleteSession(router)

		sessId := AuthenticateAdmin(app, router)

		r := AuthenticatedRequest(app, http.MethodDelete, "/api/v1/session/"+sessId, sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("UserWithoutAuthentication", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DeleteSession(router)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		r := PerformRequest(app, http.MethodDelete, "/api/v1/session/"+sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("UserAuthenticatedRequest", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DeleteSession(router)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		r := AuthenticatedRequest(app, http.MethodDelete, "/api/v1/session/"+sessId, sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidSession", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		sessId := "638bffc9b86a8fda0d908ebee84a43930cb8d1e3507f4aa0"

		DeleteSession(router)

		r := PerformRequest(app, http.MethodDelete, "/api/v1/session/"+sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
