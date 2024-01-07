package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSession(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		sess := get.Session().Public()
		assert.Equal(t, sess, Session(""))
		assert.Equal(t, sess, Session("638bffc9b86a8fda0d908ebee84a43930cb8d1e3507f4aa0"))
	})
}

func TestSessionResponse(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		sess := get.Session().Public()
		conf := get.Config().ClientSession(sess)

		// Create response in public mode.
		result := GetSessionResponse(sess.AuthToken(), sess, conf)

		// Check response.
		assert.Equal(t, StatusSuccess, result["status"])
		assert.Equal(t, sess.ID, result["session_id"])
		assert.Equal(t, sess.AuthToken(), result["id"])
		assert.Equal(t, sess.AuthToken(), result["access_token"])
		assert.Equal(t, sess.AuthTokenType(), result["token_type"])
		assert.Equal(t, sess.ExpiresIn(), result["expires_in"])
		assert.Equal(t, sess.Provider().String(), result["provider"])
		assert.Equal(t, sess.User(), result["user"])
		assert.Equal(t, sess.Data(), result["data"])
		assert.Equal(t, conf, result["config"])
	})
	t.Run("NoAuthToken", func(t *testing.T) {
		sess := get.Session().Public()
		conf := get.Config().ClientSession(sess)

		// Create response without auth token.
		result := GetSessionResponse("", sess, conf)

		// Check response.
		assert.Equal(t, StatusSuccess, result["status"])
		assert.Equal(t, sess.ID, result["session_id"])
		assert.Nil(t, result["id"])
		assert.Nil(t, result["access_token"])
		assert.Nil(t, result["token_type"])
		assert.Equal(t, sess.ExpiresIn(), result["expires_in"])
		assert.Equal(t, sess.Provider().String(), result["provider"])
		assert.Equal(t, sess.User(), result["user"])
		assert.Equal(t, sess.Data(), result["data"])
		assert.Equal(t, conf, result["config"])
	})
}

func TestCreateSession(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		CreateSession(router)

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "admin", "password": "photoprism"}`)
		t.Logf("Response Body: %s", r.Body.String())
		userName := gjson.Get(r.Body.String(), "user.Name").String()
		assert.Equal(t, "admin", userName)
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

		authToken := AuthenticateUser(app, router, "alice", "Alice123!")

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"token": "xxx"}`, authToken)
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

		authToken := AuthenticateUser(app, router, "alice", "Alice123!")

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"token": "1jxf3jfn2k"}`, authToken)
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
		authToken := AuthenticateAdmin(app, router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/session/"+rnd.SessionID(authToken))
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("AdminAuthenticatedRequest", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetSession(router)
		authToken := AuthenticateAdmin(app, router)

		t.Logf("Session ID: %s", authToken)
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/session/"+rnd.SessionID(authToken), authToken)
		t.Logf("Response Body: %s", r.Body.String())
		id := gjson.Get(r.Body.String(), "session_id").String()
		assert.Equal(t, rnd.SessionID(authToken), id)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestDeleteSession(t *testing.T) {
	t.Run("AdminWithoutAuthentication", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DeleteSession(router)
		authToken := AuthenticateAdmin(app, router)

		// f9ae12e95a01bcc7faae6497124cd721eaf13c1dad301dbc
		t.Logf("authToken: %s", authToken)

		r := AuthenticatedRequest(app, http.MethodDelete, "/api/v1/session/"+rnd.SessionID(authToken), authToken)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("AdminAuthenticatedRequest", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DeleteSession(router)
		authToken := AuthenticateAdmin(app, router)

		r := AuthenticatedRequest(app, http.MethodDelete, "/api/v1/session/"+rnd.SessionID(authToken), authToken)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("UserWithoutAuthentication", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DeleteSession(router)
		bobToken := "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"

		r := PerformRequest(app, http.MethodDelete, "/api/v1/session/"+rnd.SessionID(bobToken))
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("UserAuthenticatedRequest", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DeleteSession(router)
		authToken := AuthenticateUser(app, router, "alice", "Alice123!")

		r := AuthenticatedRequest(app, http.MethodDelete, "/api/v1/session/"+rnd.SessionID(authToken), authToken)
		t.Logf("Response Body: %s", r.Body.String())
		id := gjson.Get(r.Body.String(), "session_id").String()
		assert.Equal(t, rnd.SessionID(authToken), id)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("AliceSessionAsBob", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DeleteSession(router)
		bobToken := AuthenticateUser(app, router, "bob", "Bobbob123!")
		aliceToken := "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"

		r := AuthenticatedRequest(app, http.MethodDelete, "/api/v1/session/"+rnd.SessionID(aliceToken), bobToken)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("BobSessionAsAlice", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DeleteSession(router)
		aliceToken := AuthenticateUser(app, router, "alice", "Alice123!")
		bobToken := "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"

		r := AuthenticatedRequest(app, http.MethodDelete, "/api/v1/session/"+rnd.SessionID(bobToken), aliceToken)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("InvalidSession", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		DeleteSession(router)
		authToken := AuthenticateUser(app, router, "alice", "Alice123!")
		deleteToken := "638bffc9b86a8fda0d908ebee84a43930cb8d1e3507f4aa0"

		r := AuthenticatedRequest(app, http.MethodDelete, "/api/v1/session/"+rnd.SessionID(deleteToken), authToken)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
