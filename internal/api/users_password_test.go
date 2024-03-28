package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/form"
)

func TestChangePassword(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateUserPassword(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/users/xxx/password", `{}`)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)
		sessId := AuthenticateUser(app, router, "jens.mander", "Alice123!")

		f := form.ChangePassword{
			OldPassword: "Alice123!",
			NewPassword: "aliceinwonderland",
		}
		if pwStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxetse3cy5eo9z2/password",
				string(pwStr), sessId)
			assert.Equal(t, http.StatusUnauthorized, r.Code)
		}
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxetse3cy5eo9z2/password",
			"{OldPassword: old}", sessId)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

	t.Run("AliceProvidesWrongPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.ChangePassword{
			OldPassword: "someonewhoisntalice",
			NewPassword: "aliceinwonderland",
		}
		if pwStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxetse3cy5eo9z2/password",
				string(pwStr), sessId)
			assert.Equal(t, http.StatusBadRequest, r.Code)
		}
	})

	t.Run("Ok", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)

		oldPassword := "PleaseChange$42"
		newPassword := "SoftwareDevelopmentIsAYoungProfession1234567890!@#$%^&*()_+[]{}|:<>?/.,"

		sessId := AuthenticateUser(app, router, "fowler", oldPassword)

		frm := form.ChangePassword{
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}

		if jsonFrm, err := json.Marshal(frm); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/urinotv3d6jedvlm/password",
				string(jsonFrm), sessId)
			assert.Equal(t, http.StatusOK, r.Code)
		}

		frm = form.ChangePassword{
			OldPassword: newPassword,
			NewPassword: oldPassword,
		}

		if jsonFrm, err := json.Marshal(frm); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/urinotv3d6jedvlm/password",
				string(jsonFrm), sessId)
			assert.Equal(t, http.StatusOK, r.Code)
		}
	})

	t.Run("AliceChangesOtherUsersPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.ChangePassword{
			OldPassword: "Bobbob123!",
			NewPassword: "helloworld",
		}
		if pwStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxc08w3d0ej2283/password",
				string(pwStr), sessId)
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})

	t.Run("BobProvidesWrongPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)
		sessId := AuthenticateUser(app, router, "bob", "Bobbob123!")

		f := form.ChangePassword{
			OldPassword: "helloworld",
			NewPassword: "Bobbob123!",
		}
		if pwStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxc08w3d0ej2283/password",
				string(pwStr), sessId)
			assert.Equal(t, http.StatusBadRequest, r.Code)
		}
	})

	t.Run("SameNewPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)
		sessId := AuthenticateUser(app, router, "friend", "!Friend321")

		f := form.ChangePassword{
			OldPassword: "!Friend321",
			NewPassword: "!Friend321",
		}
		if pwStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxqg7i1kperxvu7/password",
				string(pwStr), sessId)
			assert.Equal(t, http.StatusOK, r.Code)
		}
	})

	t.Run("BobChangesOtherUsersPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)
		sessId := AuthenticateUser(app, router, "bob", "Bobbob123!")

		f := form.ChangePassword{
			OldPassword: "aliceinwonderland",
			NewPassword: "bobinwonderland",
		}
		if pwStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxetse3cy5eo9z2/password",
				string(pwStr), sessId)
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})

}
