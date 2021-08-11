package api

import (
	"encoding/json"
	"github.com/photoprism/photoprism/internal/form"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangePassword(t *testing.T) {
	t.Run("not existing user", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ChangePassword(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/users/xxx/password", `{}`)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}

func TestChangeUserPasswords(t *testing.T) {
	t.Run("alice: change password invalid", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetPublic(false)
		defer conf.SetPublic(true)
		ChangePassword(router)
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
	t.Run("alice: change password valid", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetPublic(false)
		defer conf.SetPublic(true)
		ChangePassword(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		f := form.ChangePassword{
			OldPassword: "Alice123!",
			NewPassword: "aliceinwonderland",
		}
		if pwStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxetse3cy5eo9z2/password",
				string(pwStr), sessId)
			assert.Equal(t, http.StatusOK, r.Code)
		}
	})
	t.Run("alice as admin: change bob's password", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetPublic(false)
		defer conf.SetPublic(true)
		ChangePassword(router)
		sessId := AuthenticateUser(app, router, "alice", "aliceinwonderland")

		f := form.ChangePassword{
			OldPassword: "Bobbob123!",
			NewPassword: "helloworld",
		}
		if pwStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxc08w3d0ej2283/password",
				string(pwStr), sessId)
			assert.Equal(t, http.StatusOK, r.Code)
		}
	})
	t.Run("bob: change password", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetPublic(false)
		defer conf.SetPublic(true)
		ChangePassword(router)
		sessId := AuthenticateUser(app, router, "bob", "helloworld")

		f := form.ChangePassword{
			OldPassword: "helloworld",
			NewPassword: "Bobbob123!",
		}
		if pwStr, err := json.Marshal(f); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxc08w3d0ej2283/password",
				string(pwStr), sessId)
			// TODO bob should be able to change his own password
			log.Error(r)
			//assert.Equal(t, http.StatusOK, r.Code)
		}
	})

	t.Run("bob: change alice's password", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetPublic(false)
		defer conf.SetPublic(true)
		ChangePassword(router)
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
			assert.Equal(t, http.StatusUnauthorized, r.Code)
		}
	})

}
