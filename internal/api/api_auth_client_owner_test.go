package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newOwnedClientSession creates a user, a client that belongs to it, and a stored client session.
func newOwnedClientSession(t *testing.T, scope string) (*entity.User, *entity.Session) {
	t.Helper()

	user := entity.NewUser()
	user.UserName = "client-owner-" + rnd.Base36(6)
	user.UserRole = acl.RoleAdmin.String()
	user.CanLogin = true
	require.NoError(t, user.Create())

	client := entity.NewClient().SetName("Owned Client").SetRole(acl.RoleClient.String()).SetUser(user)
	client.AuthScope = scope
	require.NoError(t, client.Create())

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/oauth/token", nil)
	c.Request.RemoteAddr = "10.4.4.4:1234"

	sess := client.NewSession(c, authn.GrantClientCredentials)
	require.NoError(t, sess.Save())

	t.Cleanup(func() {
		_ = sess.Delete()
		entity.UnscopedDb().Unscoped().Delete(client)
		entity.UnscopedDb().Unscoped().Delete(&entity.UserDetails{}, "user_uid = ?", user.UserUID)
		entity.UnscopedDb().Unscoped().Delete(&entity.UserSettings{}, "user_uid = ?", user.UserUID)
		entity.UnscopedDb().Unscoped().Delete(user)
	})

	return user, sess
}

// disableStoredClient sets the stored auth_enabled value of the client to false and flushes the client cache.
func disableStoredClient(clientUID string) error {
	err := entity.UnscopedDb().Model(&entity.Client{}).Where("client_uid = ?", clientUID).UpdateColumn("auth_enabled", false).Error
	entity.FlushClientCache()

	return err
}

// authStatus returns the HTTP status of an authorization check with the given token.
func authStatus(token string, resource acl.Resource, perm acl.Permission) int {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/"+resource.String(), nil)
	c.Request.RemoteAddr = "10.4.4.4:1234"
	header.SetAuthorization(c.Request, token)

	return Auth(c, resource, perm).HttpStatus()
}

// TestAuthAny_UserBoundClient checks how the owner's account state applies to a client that belongs to it.
func TestAuthAny_UserBoundClient(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	CreateSession(router)

	t.Run("ProviderNone", func(t *testing.T) {
		user, sess := newOwnedClientSession(t, "metrics")
		token := sess.AuthToken()

		require.Equal(t, http.StatusOK, authStatus(token, acl.ResourceMetrics, acl.AccessAll))

		require.NoError(t, user.SetProvider(authn.ProviderNone).Save())
		assert.Equal(t, http.StatusForbidden, authStatus(token, acl.ResourceMetrics, acl.AccessAll))

		user.AuthProvider = ""
		require.NoError(t, user.Save())
		assert.Equal(t, http.StatusOK, authStatus(token, acl.ResourceMetrics, acl.AccessAll))
	})
	t.Run("WebLoginDisabled", func(t *testing.T) {
		user, sess := newOwnedClientSession(t, "metrics")
		token := sess.AuthToken()

		user.CanLogin = false
		require.NoError(t, user.Save())

		// The client keeps working within its own scope.
		assert.Equal(t, http.StatusOK, authStatus(token, acl.ResourceMetrics, acl.AccessAll))
		assert.Equal(t, http.StatusForbidden, authStatus(token, acl.ResourcePhotos, acl.ActionView))

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/session", `{}`, token)
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
}

func TestAuthAny_UserBoundClientOwnerGone(t *testing.T) {
	_, _, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	user, sess := newOwnedClientSession(t, "metrics")
	token := sess.AuthToken()
	require.Equal(t, http.StatusOK, authStatus(token, acl.ResourceMetrics, acl.AccessAll))

	// A client whose owner no longer exists is refused on every request.
	require.NoError(t, entity.UnscopedDb().Unscoped().Delete(user).Error)
	entity.FlushSessionCache()
	assert.Equal(t, http.StatusForbidden, authStatus(token, acl.ResourceMetrics, acl.AccessAll))
	assert.Equal(t, http.StatusForbidden, authStatus(token, acl.ResourceMetrics, acl.AccessAll))
}

func TestSessionReusable(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, sessionReusable(nil))
	})
	t.Run("ClientWithoutSessionScope", func(t *testing.T) {
		_, sess := newOwnedClientSession(t, "metrics")
		assert.False(t, sessionReusable(sess))
	})
	t.Run("ClientWithSessionScope", func(t *testing.T) {
		_, sess := newOwnedClientSession(t, "*")
		assert.True(t, sessionReusable(sess))
	})
	t.Run("AppPasswordOfDeniedOwner", func(t *testing.T) {
		owner := *entity.UserFixtures.Pointer("bob")
		owner.CanLogin = false
		sess, err := entity.AddClientSession("reuse-app-password", 3600, "*", authn.GrantPassword, &owner)
		require.NoError(t, err)
		t.Cleanup(func() { _ = sess.Delete() })
		require.True(t, sess.IsApplication())

		assert.False(t, sessionReusable(sess))
	})
	t.Run("ClientDisabled", func(t *testing.T) {
		_, sess := newOwnedClientSession(t, "*")
		require.True(t, sessionReusable(sess))
		require.NoError(t, disableStoredClient(sess.ClientUID))
		assert.False(t, sessionReusable(sess))
	})
	t.Run("UserSession", func(t *testing.T) {
		sess := entity.NewSession(3600, 0).SetUser(entity.UserFixtures.Pointer("alice"))
		require.NoError(t, sess.Save())
		t.Cleanup(func() { _ = sess.Delete() })

		assert.True(t, sessionReusable(sess))
	})
	t.Run("RemovedRow", func(t *testing.T) {
		sess := entity.NewSession(3600, 0).SetUser(entity.UserFixtures.Pointer("alice"))
		require.NoError(t, sess.Save())
		require.NoError(t, entity.UnscopedDb().Exec("DELETE FROM auth_sessions WHERE id = ?", sess.ID).Error)

		assert.False(t, sessionReusable(sess))
	})
}

func TestAuthorizeSession(t *testing.T) {
	t.Run("UserSession", func(t *testing.T) {
		sess := entity.NewSession(3600, 0).SetUser(entity.UserFixtures.Pointer("alice"))
		assert.True(t, authorizeSession("10.4.4.4", sess, acl.ResourceConfig, acl.Permissions{acl.ActionView}).Valid())
	})
	t.Run("ClientOutsideScope", func(t *testing.T) {
		_, sess := newOwnedClientSession(t, "metrics")
		assert.False(t, authorizeSession("10.4.4.4", sess, acl.ResourceConfig, acl.Permissions{acl.ActionView}).Valid())
	})
	t.Run("ClientDisabled", func(t *testing.T) {
		_, sess := newOwnedClientSession(t, "*")
		require.True(t, authorizeSession("10.4.4.4", sess, acl.ResourceConfig, acl.Permissions{acl.ActionView}).Valid())
		require.NoError(t, disableStoredClient(sess.ClientUID))

		result := authorizeSession("10.4.4.4", sess, acl.ResourceConfig, acl.Permissions{acl.ActionView})
		assert.Equal(t, http.StatusUnauthorized, result.HttpStatus())
	})
	t.Run("ClientDisabledOtherProvider", func(t *testing.T) {
		_, clientSess := newOwnedClientSession(t, "*")
		require.NoError(t, disableStoredClient(clientSess.ClientUID))

		sess := entity.NewSession(3600, 0).SetUser(entity.UserFixtures.Pointer("alice"))
		sess.ClientUID = clientSess.ClientUID
		require.False(t, sess.IsClient())

		result := authorizeSession("10.4.4.4", sess, acl.ResourceConfig, acl.Permissions{acl.ActionView})
		assert.Equal(t, http.StatusUnauthorized, result.HttpStatus())
	})
	t.Run("ClientOfDeactivatedOwner", func(t *testing.T) {
		user, sess := newOwnedClientSession(t, "*")
		require.NoError(t, user.SetProvider(authn.ProviderNone).Save())

		result := authorizeSession("10.4.4.4", sess, acl.ResourceConfig, acl.Permissions{acl.ActionView})
		assert.Equal(t, http.StatusForbidden, result.HttpStatus())
	})
}

// TestAuthAny_UserBoundClientDeactivatedOwner checks that a client of a deactivated owner is refused on sign-in
// and downloads.
func TestAuthAny_UserBoundClientDeactivatedOwner(t *testing.T) {
	app, router, conf := NewApiTest()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options; conf.Propagate() })
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })
	conf.Options().OriginalsPath = t.TempDir()
	conf.Propagate()

	CreateSession(router)
	GetClientConfig(router)
	GetPhotoDownload(router)

	user, sess := newOwnedClientSession(t, "*")
	token := sess.AuthToken()

	const uid = "ps6sg6be2lvl0y13"
	file, err := query.FileByPhotoUID(uid)
	require.NoError(t, err)
	CreateTestOriginal(t, file)

	cfg := AuthenticatedRequest(app, http.MethodGet, "/api/v1/config", token)
	require.Equal(t, http.StatusOK, cfg.Code)
	downloadToken := gjson.GetBytes(cfg.Body.Bytes(), "downloadToken").String()
	require.NotEmpty(t, downloadToken)
	require.Equal(t, http.StatusOK, PerformRequest(app, http.MethodGet, "/api/v1/photos/"+uid+"/dl?t="+downloadToken).Code)
	require.True(t, sessionReusable(sess))

	require.NoError(t, user.SetProvider(authn.ProviderNone).Save())

	assert.False(t, sessionReusable(sess))
	assert.Equal(t, http.StatusUnauthorized, AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/session", `{}`, token).Code)
	assert.Equal(t, http.StatusForbidden, PerformRequest(app, http.MethodGet, "/api/v1/photos/"+uid+"/dl?t="+downloadToken).Code)
}

// failOwnerReads makes the next n queries of the users table fail and returns the number of queries seen.
func failOwnerReads(t *testing.T, n int32) *atomic.Int32 {
	t.Helper()

	var seen atomic.Int32
	name := "test:fail_owner_reads_" + rnd.Base36(8)
	require.NoError(t, entity.Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
		if db.Statement.Table == (entity.User{}).TableName() && seen.Add(1) <= n {
			_ = db.AddError(errors.New("owner lookup failed"))
		}
	}))
	t.Cleanup(func() { _ = entity.Db().Callback().Query().Remove(name) })

	return &seen
}

// loadOwnedClientSession returns a client session with full access whose owner is limited to the photos scope,
// loaded while the first read of the owner fails.
func loadOwnedClientSession(t *testing.T, failReads int32) *entity.Session {
	t.Helper()

	user, sess := newOwnedClientSession(t, "*")
	user.UserScope = "photos"
	require.NoError(t, user.Save())
	entity.FlushSessionCache()
	entity.FlushClientCache()

	seen := failOwnerReads(t, failReads)
	s, err := entity.FindAnySession(sess.ID)
	require.NoError(t, err)
	require.Equal(t, int32(1), seen.Load())
	require.Equal(t, "*", s.Scope())

	return s
}

func TestAuthorizeSession_OwnerScopeBeforeScopeCheck(t *testing.T) {
	_, _, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	t.Run("OwnerReadSucceeds", func(t *testing.T) {
		s := loadOwnedClientSession(t, 1)
		assert.Equal(t, http.StatusForbidden, authorizeSession("10.4.4.4", s, acl.ResourceMetrics, acl.Permissions{acl.AccessAll}).HttpStatus())
	})
	t.Run("OwnerReadFails", func(t *testing.T) {
		// The read during authorization fails as well, and a later read would succeed.
		s := loadOwnedClientSession(t, 2)
		assert.Equal(t, http.StatusForbidden, authorizeSession("10.4.4.4", s, acl.ResourceMetrics, acl.Permissions{acl.AccessAll}).HttpStatus())
	})
	t.Run("NoOwner", func(t *testing.T) {
		sess, err := entity.AddClientSession("owner-less-client", 3600, "metrics", authn.GrantClientCredentials, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = sess.Delete() })
		require.True(t, sess.NoUser())

		seen := failOwnerReads(t, 0)
		assert.True(t, authorizeSession("10.4.4.4", sess, acl.ResourceMetrics, acl.Permissions{acl.AccessAll}).Valid())
		assert.Equal(t, int32(0), seen.Load())
	})
}

func TestSessionReusable_OwnerScope(t *testing.T) {
	_, _, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	t.Run("OwnerReadSucceeds", func(t *testing.T) {
		s := loadOwnedClientSession(t, 1)
		assert.False(t, sessionReusable(s))
	})
	t.Run("OwnerReadFails", func(t *testing.T) {
		s := loadOwnedClientSession(t, 2)
		assert.False(t, sessionReusable(s))
	})
	t.Run("AppPasswordOwnerReadFails", func(t *testing.T) {
		sess, err := entity.AddClientSession("reuse-app-password", 3600, "*", authn.GrantPassword, entity.UserFixtures.Pointer("bob"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = sess.Delete() })
		require.True(t, sess.IsApplication())
		entity.FlushSessionCache()

		failOwnerReads(t, 100)
		s, err := entity.FindAnySession(sess.ID)
		require.NoError(t, err)
		assert.False(t, sessionReusable(s))
	})
}
