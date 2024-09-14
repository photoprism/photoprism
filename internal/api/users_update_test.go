package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestUpdateUser(t *testing.T) {
	t.Run("InvalidRequestBody", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s", adminUid)
		t.Logf("Request URL: %s", reqUrl)
		r := AuthenticatedRequestWithBody(app, "PUT", reqUrl, "{Email:\"admin@example.com\",Details:{Location:\"WebStorm\"}}", sessId)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s", adminUid)
		UpdateUser(router)
		r := PerformRequestWithBody(app, "PUT", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)
		sessId := AuthenticateUser(app, router, "jens.mander", "Alice123!")

		f := form.User{
			DisplayName: "New Name",
		}

		if userForm, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxetse3cy5eo9z2",
				string(userForm), sessId)
			assert.Equal(t, http.StatusUnauthorized, r.Code)
		}
	})

	t.Run("AliceChangeOwn", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.User{
			DisplayName: "Alicia",
			UploadPath:  "uploads-alice",
		}

		if userForm, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxetse3cy5eo9z2",
				string(userForm), sessId)
			assert.Equal(t, http.StatusOK, r.Code)
			assert.Contains(t, r.Body.String(), "\"DisplayName\":\"Alicia\"")
			assert.Contains(t, r.Body.String(), "\"UploadPath\":\"uploads-alice\"")
		}
	})

	t.Run("AliceChangeBob", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.User{
			DisplayName: "Bobby",
			WebDAV:      false,
			UploadPath:  "uploads-bob",
		}

		if userForm, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxc08w3d0ej2283",
				string(userForm), sessId)
			assert.Equal(t, http.StatusOK, r.Code)
			assert.Contains(t, r.Body.String(), "\"DisplayName\":\"Bobby\"")
			assert.Contains(t, r.Body.String(), "\"UploadPath\":\"uploads-bob\"")
		}
	})

	t.Run("BobChangeOwn", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)
		sessId := AuthenticateUser(app, router, "bob", "Bobbob123!")

		f := form.User{
			DisplayName: "Bobo",
		}

		if userForm, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxc08w3d0ej2283",
				string(userForm), sessId)
			assert.Equal(t, http.StatusOK, r.Code)
			assert.Contains(t, r.Body.String(), "\"DisplayName\":\"Bobo\"")
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.User{
			DisplayName: "Bobby",
		}

		if userForm, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxc08w3d0ej2555",
				string(userForm), sessId)
			assert.Equal(t, http.StatusNotFound, r.Code)
		}
	})
}
