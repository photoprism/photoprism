package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/form"

	"github.com/stretchr/testify/assert"
)

func TestCreateUserPasscode(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CreateUserPasscode(router)
		r := PerformRequest(app, "POST", "/api/v1/users/uqxc08w3d0ej2283/passcode")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("Unauthorized", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		CreateUserPasscode(router)

		r := PerformRequest(app, "POST", "/api/v1/users/uqxc08w3d0ej2283/passcode")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("UsersDontMatch", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		CreateUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "",
			Password: "Alice123!",
			Type:     "totp",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxc08w3d0ej2283/passcode", string(pcStr), sessId)
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})
	t.Run("AliceUnsupportedType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		CreateUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "",
			Password: "abcdef",
			Type:     "xxx",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode", string(pcStr), sessId)
			assert.Equal(t, http.StatusBadRequest, r.Code)
		}
	})
	t.Run("AliceInvalidPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		CreateUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "",
			Password: "wrong",
			Type:     "totp",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode", string(pcStr), sessId)
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})
	t.Run("AliceSuccess", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		CreateUserPasscode(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.UserPasscode{
			Passcode: "",
			Password: "Alice123!",
			Type:     "totp",
		}
		if pcStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "POST", "/api/v1/users/uqxetse3cy5eo9z2/passcode", string(pcStr), sessId)
			assert.Equal(t, http.StatusOK, r.Code)
		}
	})
}
