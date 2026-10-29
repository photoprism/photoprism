package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
	"github.com/photoprism/photoprism/pkg/txt"
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
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		body := `{"OldPassword":"Alice123!","NewPassword":"` + strings.Repeat("a", int(MaxAuthRequestBytes)) + `"}`
		r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxetse3cy5eo9z2/password", body, sessId)

		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
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
	t.Run("Success", func(t *testing.T) {
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
	t.Run("RotatesPreviewToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)

		const oldPassword = "RotatePreview123!"
		u := entity.NewUser()
		u.UserName = "rotate-preview"
		u.UserRole = acl.RoleAdmin.String()
		u.AuthProvider = authn.ProviderLocal.String()
		u.CanLogin = true
		require.NoError(t, entity.Db().Create(u).Error)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Session{}, "user_uid = ?", u.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Password{}, "uid = ?", u.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserDetails{}, "user_uid = ?", u.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserSettings{}, "user_uid = ?", u.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.User{}, "user_uid = ?", u.UserUID).Error)
		})
		require.NoError(t, u.SetPassword(oldPassword))
		oldPreview := u.PreviewToken

		sessId := AuthenticateUser(app, router, u.UserName, oldPassword)
		other := entity.NewSession(unix.Day, unix.Hour)
		other.SetUser(u)
		require.NoError(t, other.Save())
		otherId := other.AuthToken()

		// App passwords outlive a password change, so their cached sessions must pick up the new token.
		appPw, err := entity.AddClientSession("Rotate Preview Client", unix.Day, "*", authn.GrantPassword, u)
		require.NoError(t, err)
		cached, err := entity.FindSession(appPw.ID)
		require.NoError(t, err)
		require.Equal(t, oldPreview, cached.PreviewToken)

		frm := form.ChangePassword{OldPassword: oldPassword, NewPassword: "RotatePreview456!"}
		jsonFrm, err := json.Marshal(frm)
		require.NoError(t, err)

		r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/"+u.UserUID+"/password", string(jsonFrm), sessId)
		require.Equal(t, http.StatusOK, r.Code)

		updated := entity.FindUserByUID(u.UserUID)
		require.NotNil(t, updated)
		require.NotEqual(t, oldPreview, updated.PreviewToken)
		assert.Equal(t, updated.PreviewToken, r.Header().Get("X-Preview-Token"))
		assert.True(t, entity.InvalidPreviewToken(oldPreview))
		assert.False(t, entity.InvalidPreviewToken(updated.PreviewToken))

		found, err := entity.FindSession(appPw.ID)
		require.NoError(t, err)
		assert.Equal(t, updated.PreviewToken, found.PreviewToken)

		// Other login sessions are revoked.
		_, err = entity.FindSession(rnd.SessionID(otherId))
		assert.Error(t, err)

		// The session that changed the password keeps working.
		frm = form.ChangePassword{OldPassword: "RotatePreview456!", NewPassword: oldPassword}
		jsonFrm, err = json.Marshal(frm)
		require.NoError(t, err)

		r = AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/"+u.UserUID+"/password", string(jsonFrm), sessId)
		require.Equal(t, http.StatusOK, r.Code)
		assert.NotEqual(t, updated.PreviewToken, r.Header().Get("X-Preview-Token"))
		assert.False(t, entity.InvalidPreviewToken(r.Header().Get("X-Preview-Token")))
	})
	t.Run("InvalidNewPasswordKeepsSessions", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)

		const oldPassword = "KeepSessions123!"
		u := entity.NewUser()
		u.UserName = "keep-sessions"
		u.UserRole = acl.RoleAdmin.String()
		u.AuthProvider = authn.ProviderLocal.String()
		u.CanLogin = true
		require.NoError(t, entity.Db().Create(u).Error)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Session{}, "user_uid = ?", u.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Password{}, "uid = ?", u.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserDetails{}, "user_uid = ?", u.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserSettings{}, "user_uid = ?", u.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.User{}, "user_uid = ?", u.UserUID).Error)
		})
		require.NoError(t, u.SetPassword(oldPassword))

		sessId := AuthenticateUser(app, router, u.UserName, oldPassword)
		other := entity.NewSession(unix.Day, unix.Hour)
		other.SetUser(u)
		require.NoError(t, other.Save())
		otherId := other.AuthToken()

		jsonFrm, err := json.Marshal(form.ChangePassword{OldPassword: oldPassword, NewPassword: strings.Repeat("a", txt.ClipPassword+1)})
		require.NoError(t, err)

		r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/"+u.UserUID+"/password", string(jsonFrm), sessId)
		assert.Equal(t, http.StatusBadRequest, r.Code)

		_, err = entity.FindSession(rnd.SessionID(otherId))
		assert.NoError(t, err)
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
	t.Run("AliceAppPassword", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)

		oldPassword := "PleaseChange$42"
		newPassword := "SoftwareDevelopmentIsAYoungProfession1234567890!@#$%^&*()_+[]{}|:<>?/.,"

		frm := form.ChangePassword{
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}

		if jsonFrm, err := json.Marshal(frm); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/urinotv3d6jedvlm/password",
				string(jsonFrm), "X3B6IU-hfeLG5-HpVxkT-ctCY3M")
			assert.Equal(t, http.StatusForbidden, r.Code)
			val := gjson.Get(r.Body.String(), "error")
			assert.Equal(t, "Permission denied", val.String())
		}
	})
	t.Run("AliceAppPasswordWebdav", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)

		oldPassword := "PleaseChange$42"
		newPassword := "SoftwareDevelopmentIsAYoungProfession1234567890!@#$%^&*()_+[]{}|:<>?/.,"

		frm := form.ChangePassword{
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}

		if jsonFrm, err := json.Marshal(frm); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/urinotv3d6jedvlm/password",
				string(jsonFrm), "v2wS72-OkqEzm-MQ63Z2-TEhU0w")
			assert.Equal(t, http.StatusForbidden, r.Code)
			val := gjson.Get(r.Body.String(), "error")
			assert.Equal(t, "Permission denied", val.String())
		}
	})
	t.Run("AccessToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUserPassword(router)

		oldPassword := "PleaseChange$42"
		newPassword := "SoftwareDevelopmentIsAYoungProfession1234567890!@#$%^&*()_+[]{}|:<>?/.,"

		frm := form.ChangePassword{
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}

		if jsonFrm, err := json.Marshal(frm); err != nil {
			log.Fatal(err)
		} else {
			r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/urinotv3d6jedvlm/password",
				string(jsonFrm), "8e154d323800393faf5177ce7392116feebbf674e6c2d39e")
			assert.Equal(t, http.StatusForbidden, r.Code)
			val := gjson.Get(r.Body.String(), "error")
			assert.Equal(t, "Permission denied", val.String())
		}
	})

}
