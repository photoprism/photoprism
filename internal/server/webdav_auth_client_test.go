package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestWebDAVAuth_UserBoundClient checks WebDAV access with the token of a client whose owner cannot use web login.
func TestWebDAVAuth_UserBoundClient(t *testing.T) {
	conf := config.TestConfig()
	webdavHandler := WebDAVAuth(conf)

	// newClientToken creates an owner with web login disabled, a client bound to it, and returns a token.
	newClientToken := func(t *testing.T, scope string, owner func(u *entity.User)) string {
		t.Helper()

		user := entity.NewUser()
		user.UserName = "webdav-owner-" + rnd.Base36(6)
		user.UserRole = acl.RoleAdmin.String()
		user.CanLogin = false
		user.WebDAV = true
		require.NoError(t, user.Create())

		client := entity.NewClient().SetName("WebDAV Client").SetRole(acl.RoleClient.String()).SetUser(user)
		client.AuthScope = scope
		require.NoError(t, client.Create())

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/oauth/token", nil)
		c.Request.RemoteAddr = "10.5.5.5:1234"

		sess := client.NewSession(c, authn.GrantClientCredentials)
		require.NoError(t, sess.Save())

		if owner != nil {
			owner(user)
		}

		t.Cleanup(func() {
			_ = sess.Delete()
			entity.UnscopedDb().Unscoped().Delete(client)
			entity.UnscopedDb().Unscoped().Delete(&entity.UserDetails{}, "user_uid = ?", user.UserUID)
			entity.UnscopedDb().Unscoped().Delete(&entity.UserSettings{}, "user_uid = ?", user.UserUID)
			entity.UnscopedDb().Unscoped().Delete(user)
		})

		return sess.AuthToken()
	}

	request := func(token string) int {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = &http.Request{Method: http.MethodGet, Header: make(http.Header)}
		header.SetAuthorization(c.Request, token)

		entity.FlushSessionCache()
		webdavHandler(c)

		return c.Writer.Status()
	}

	t.Run("Allowed", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, request(newClientToken(t, "webdav", nil)))
	})
	t.Run("ClientWithoutWebDAVScope", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, request(newClientToken(t, "metrics", nil)))
	})
	t.Run("OwnerWithoutWebDAV", func(t *testing.T) {
		token := newClientToken(t, "webdav", func(u *entity.User) {
			require.NoError(t, entity.Db().Model(u).UpdateColumn("webdav", false).Error)
		})
		assert.Equal(t, http.StatusUnauthorized, request(token))
	})
	t.Run("OwnerDeleted", func(t *testing.T) {
		token := newClientToken(t, "webdav", func(u *entity.User) {
			deletedAt := entity.Now()
			require.NoError(t, entity.Db().Model(u).UpdateColumn("deleted_at", &deletedAt).Error)
		})
		assert.Equal(t, http.StatusUnauthorized, request(token))
	})
	t.Run("OwnerProviderNone", func(t *testing.T) {
		token := newClientToken(t, "webdav", func(u *entity.User) {
			require.NoError(t, entity.Db().Model(u).UpdateColumn("auth_provider", authn.ProviderNone.String()).Error)
		})
		assert.Equal(t, http.StatusUnauthorized, request(token))
	})
}
