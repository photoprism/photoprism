package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

func TestUpdateUser(t *testing.T) {
	t.Run("InvalidRequestBody", func(t *testing.T) {
		// Body validation runs after the ownership check, so target the
		// session user's own UID to exercise the BindJSON failure path.
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")
		aliceUid := "uqxetse3cy5eo9z2"
		reqUrl := fmt.Sprintf("/api/v1/users/%s", aliceUid)
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
		// Community Edition grants admins own-account user management only;
		// full-access editions permit cross-account admin profile updates.
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
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})
	t.Run("GuestCannotEditOtherUser", func(t *testing.T) {
		// Guest sessions may update their own profile, but not another account.
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)

		guestUsername := "guest_update_idor_test"
		if err := entity.AddUser(form.User{
			UserName: guestUsername,
			UserRole: acl.RoleGuest.String(),
			Password: "GuestPass123!",
			CanLogin: true,
		}); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if u := entity.FindUserByName(guestUsername); u != nil {
				_ = u.Delete()
			}
		})

		sessId := AuthenticateUser(app, router, guestUsername, "GuestPass123!")
		if sessId == "" {
			t.Fatal("guest authentication failed")
		}

		adminUid := entity.Admin.UserUID
		body, err := json.Marshal(form.User{
			UserEmail:   "attacker@example.test",
			DisplayName: "PWNED",
		})
		if err != nil {
			t.Fatal(err)
		}

		r := AuthenticatedRequestWithBody(app, "PUT", fmt.Sprintf("/api/v1/users/%s", adminUid), string(body), sessId)
		assert.Equal(t, http.StatusForbidden, r.Code)

		// Confirm the admin record was not mutated.
		fresh := entity.FindUserByUID(adminUid)
		if fresh == nil {
			t.Fatal("admin user not found after guest request")
		}
		assert.NotEqual(t, "attacker@example.test", fresh.UserEmail)
		assert.NotEqual(t, "PWNED", fresh.DisplayName)
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
		// Ownership is checked before lookup, so non-admin requests for
		// unknown foreign UIDs return 403 without leaking account existence.
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
			assert.Equal(t, http.StatusForbidden, r.Code)
		}
	})
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")
		body := `{"DisplayName":"` + strings.Repeat("a", int(MaxMutationRequestBytes)) + `"}`
		r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxetse3cy5eo9z2", body, sessId)

		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
	})
}

func TestUpdateUser_Guards(t *testing.T) {
	t.Run("SelfRoleChangeForbidden", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")
		body, _ := json.Marshal(form.User{UserName: "alice", UserRole: "user"})
		r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/uqxetse3cy5eo9z2", string(body), sessId)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("SystemAccountForbidden", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")
		body, _ := json.Marshal(form.User{DisplayName: "Hacked"})
		r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/"+entity.UnknownUser.UserUID, string(body), sessId)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}

func TestUpdateUser_ClusterJWT(t *testing.T) {
	// A Portal cluster JWT (GrantJwtBearer + users-manage scope) is a user-less
	// service principal that UpdateUser authorizes like an admin so the Portal can
	// sync instance user state: it bypasses the per-user owner check and applies
	// the privilege-level fields (login, WebDAV, role) the form carries.
	t.Run("ManagesAnotherUser", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "users-update-manage")
		fx.nodeConf.Options().JWTScope = "cluster users"
		get.SetConfig(fx.nodeConf)

		username := "cluster-sync-target"
		require.NoError(t, entity.AddUser(form.User{
			UserName: username,
			UserRole: acl.RoleGuest.String(),
			Password: "ClusterSync123!",
			CanLogin: false,
			WebDAV:   false,
		}))
		target := entity.FindUserByName(username)
		require.NotNil(t, target)
		require.False(t, target.CanLogin, "precondition: target login is disabled")
		require.False(t, target.WebDAV, "precondition: target WebDAV is disabled")
		t.Cleanup(func() {
			if u := entity.FindUserByName(username); u != nil {
				_ = u.Delete()
			}
		})

		app, router, _ := NewApiTest()
		UpdateUser(router)

		spec := fx.defaultClaimsSpec()
		spec.Scope = []string{"cluster", "users"}
		token := fx.issue(t, spec)

		body, err := json.Marshal(form.User{
			UserName:    username,
			DisplayName: "Synced By Portal",
			CanLogin:    true,
			WebDAV:      true,
		})
		require.NoError(t, err)

		r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/"+target.UserUID, string(body), token)
		assert.Equal(t, http.StatusOK, r.Code)

		updated := entity.FindUserByUID(target.UserUID)
		require.NotNil(t, updated)
		assert.True(t, updated.CanLogin, "cluster JWT must apply the privilege-level CanLogin change")
		assert.True(t, updated.WebDAV, "cluster JWT must apply the privilege-level WebDAV change")
		assert.Equal(t, "Synced By Portal", updated.DisplayName)
		assert.Equal(t, acl.RoleGuest.String(), updated.UserRole, "role stays unchanged when the form omits it")
	})
	t.Run("WithoutUsersScopeDenied", func(t *testing.T) {
		fx := newPortalJWTFixture(t, "users-update-deny")
		fx.nodeConf.Options().JWTScope = "cluster users"
		get.SetConfig(fx.nodeConf)

		username := "cluster-sync-denied"
		require.NoError(t, entity.AddUser(form.User{
			UserName: username,
			UserRole: acl.RoleGuest.String(),
			Password: "ClusterSync123!",
			CanLogin: false,
		}))
		target := entity.FindUserByName(username)
		require.NotNil(t, target)
		t.Cleanup(func() {
			if u := entity.FindUserByName(username); u != nil {
				_ = u.Delete()
			}
		})

		app, router, _ := NewApiTest()
		UpdateUser(router)

		spec := fx.defaultClaimsSpec()
		spec.Scope = []string{"cluster"} // no users scope → not authorized to manage users
		token := fx.issue(t, spec)

		body, _ := json.Marshal(form.User{UserName: username, CanLogin: true})
		r := AuthenticatedRequestWithBody(app, "PUT", "/api/v1/users/"+target.UserUID, string(body), token)
		assert.NotEqual(t, http.StatusOK, r.Code, "a JWT without the users scope must not manage users")

		unchanged := entity.FindUserByUID(target.UserUID)
		require.NotNil(t, unchanged)
		assert.False(t, unchanged.CanLogin, "a denied request must not enable login")
	})
}
