package server

import (
	"encoding/base64"
	"fmt"
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

func TestWebDAVAuth(t *testing.T) {
	conf := config.TestConfig()
	webdavHandler := WebDAVAuth(conf)

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		entity.FlushSessionCache()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceToken", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		entity.FlushSessionCache()

		sess := entity.SessionFixtures.Get("alice_token")
		header.SetAuthorization(c.Request, sess.AuthToken())

		webdavHandler(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceTokenWebdav", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		sess := entity.SessionFixtures.Get("alice_token_webdav")
		basicAuth := fmt.Appendf(nil, "alice:%s", sess.AuthToken())
		c.Request.Header.Add(header.Auth, fmt.Sprintf("%s %s", header.AuthBasic, base64.StdEncoding.EncodeToString(basicAuth)))

		entity.FlushSessionCache()
		webdavHandler(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceTokenWebdavWrongUsername", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		sess := entity.SessionFixtures.Get("alice_token_webdav")
		basicAuth := fmt.Appendf(nil, "bob:%s", sess.AuthToken())
		c.Request.Header.Add(header.Auth, fmt.Sprintf("%s %s", header.AuthBasic, base64.StdEncoding.EncodeToString(basicAuth)))

		entity.FlushSessionCache()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceTokenWebdavWithoutUsername", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		sess := entity.SessionFixtures.Get("alice_token_webdav")
		basicAuth := fmt.Appendf(nil, ":%s", sess.AuthToken())
		c.Request.Header.Add(header.Auth, fmt.Sprintf("%s %s", header.AuthBasic, base64.StdEncoding.EncodeToString(basicAuth)))

		entity.FlushSessionCache()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceTokenScope", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		sess := entity.SessionFixtures.Get("alice_token_scope")
		header.SetAuthorization(c.Request, sess.AuthToken())

		entity.FlushSessionCache()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("InvalidAuthToken", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		header.SetAuthorization(c.Request, rnd.AuthToken())

		entity.FlushSessionCache()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("InvalidAppPassword", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		header.SetAuthorization(c.Request, rnd.AppPassword())

		entity.FlushSessionCache()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AppPasswordsDisabled", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// A real app-password session (OIDC-only user, session grant) read from fixtures
		// so the gate's provider check is exercised without a database write.
		sess := entity.SessionFixtures.Get("alice_app_password")
		assert.True(t, sess.IsApplication())
		header.SetAuthorization(c.Request, sess.AuthToken())

		conf.Settings().Features.AppPasswords = false
		defer func() { conf.Settings().Features.AppPasswords = true }()

		entity.FlushSessionCache()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AppPasswordAuthToken", func(t *testing.T) {
		appSess, err := entity.AddClientSession("webdav-app-password", 3600, "*", authn.GrantPassword, entity.UserFixtures.Pointer("alice"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = appSess.Delete() })
		require.True(t, appSess.IsApplication())

		request := func(extraToken string) int {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{Header: make(http.Header)}

			basicAuth := fmt.Appendf(nil, "alice:%s", appSess.AuthToken())
			c.Request.Header.Add(header.Auth, fmt.Sprintf("%s %s", header.AuthBasic, base64.StdEncoding.EncodeToString(basicAuth)))

			if extraToken != "" {
				c.Request.Header.Set(header.XAuthToken, extraToken)
			}

			entity.FlushSessionCache()
			webdavHandler(c)

			return c.Writer.Status()
		}

		// App passwords authenticate through the auth token check only.
		assert.Equal(t, http.StatusOK, request(""))
		assert.Equal(t, http.StatusUnauthorized, request("x"))

		conf.Settings().Features.AppPasswords = false
		defer func() { conf.Settings().Features.AppPasswords = true }()

		assert.Equal(t, http.StatusUnauthorized, request(""))
		assert.Equal(t, http.StatusUnauthorized, request("x"))
	})
	t.Run("AppPasswordWithoutWebDAVScope", func(t *testing.T) {
		appSess, err := entity.AddClientSession("webdav-app-password-scope", 3600, "sessions", authn.GrantPassword, entity.UserFixtures.Pointer("alice"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = appSess.Delete() })

		for _, extraToken := range []string{"", "x"} {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{Header: make(http.Header)}

			basicAuth := fmt.Appendf(nil, "alice:%s", appSess.AuthToken())
			c.Request.Header.Add(header.Auth, fmt.Sprintf("%s %s", header.AuthBasic, base64.StdEncoding.EncodeToString(basicAuth)))

			if extraToken != "" {
				c.Request.Header.Set(header.XAuthToken, extraToken)
			}

			entity.FlushSessionCache()
			webdavHandler(c)

			assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		}
	})
}

// TestWebDAVAuthSession checks credential resolution with cold and warm account caches.
func TestWebDAVAuthSession(t *testing.T) {
	t.Run("AliceTokenWebdav", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		s := entity.SessionFixtures.Get("alice_token_webdav")

		// Get session with authorized user and webdav scope.
		entity.FlushSessionCache()
		sess, user, sid, cached := WebDAVAuthSession(c, s.AuthToken())

		// Check result.
		assert.NotNil(t, sess)
		assert.NotNil(t, user)
		assert.True(t, sess.HasUser())
		assert.Equal(t, user.UserUID, sess.UserUID)
		assert.Equal(t, entity.UserFixtures.Get("alice").UserUID, sess.UserUID)
		assert.True(t, sess.ValidateScope(acl.ResourceWebDAV, acl.Permissions{acl.ActionView}))
		assert.False(t, cached)

		assert.Equal(t, s.ID, sid)
		assert.Equal(t, entity.UserFixtures.Get("alice").UserUID, user.UserUID)
		assert.True(t, user.CanUseWebDAV())

		// WebDAVAuthSession should not set a status code or any headers.
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))

		// Cache authentication.
		entity.CacheWebDAVUser(sid, user, entity.CurrentAuthCacheGeneration())

		// Get cached user.
		sess, user, sid, cached = WebDAVAuthSession(c, s.AuthToken())

		// Check result.
		assert.NotNil(t, sess)
		assert.NotNil(t, user)
		assert.True(t, cached)

		assert.Equal(t, s.ID, sid)
		assert.Equal(t, entity.UserFixtures.Get("alice").UserUID, user.UserUID)
		assert.True(t, user.CanUseWebDAV())

		// WebDAVAuthSession should not set a status code or any headers.
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceTokenScope", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		s := entity.SessionFixtures.Get("alice_token_scope")

		// Get session without sufficient authorization scope.
		sess, user, sid, cached := WebDAVAuthSession(c, s.AuthToken())

		// Check result.
		assert.NotNil(t, sess)
		assert.NotNil(t, user)
		assert.Equal(t, s.ID, sid)
		assert.False(t, cached)
		assert.True(t, sess.HasUser())
		assert.Equal(t, user.UserUID, sess.UserUID)
		assert.Equal(t, entity.UserFixtures.Get("alice").UserUID, user.UserUID)
		assert.Equal(t, entity.UserFixtures.Get("alice").UserUID, sess.UserUID)
		assert.True(t, user.CanUseWebDAV())
		assert.False(t, sess.ValidateScope(acl.ResourceWebDAV, acl.Permissions{acl.ActionView}))

		// WebDAVAuthSession should not set a status code or any headers.
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("InvalidAppPassword", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		appPassword := rnd.AppPassword()
		authId := rnd.SessionID(appPassword)

		// Get session with invalid app password.
		sess, user, sid, cached := WebDAVAuthSession(c, appPassword)

		// Check result.
		assert.Nil(t, sess)
		assert.Nil(t, user)
		assert.Equal(t, authId, sid)
		assert.False(t, cached)

		// WebDAVAuthSession should not set a status code or any headers.
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
	})
}

// TestSetWebDAVUser checks the effective account/client ceiling carried into file operations.
func TestSetWebDAVUser(t *testing.T) {
	for _, tc := range []struct {
		name, role, user string
		allowed          bool
	}{
		{"BasicAdmin", "", "alice", true}, {"BasicGuest", "", "guest", false},
		{"ClientCeiling", "instance", "alice", false}, {"AccountCeiling", "client", "guest", false}, {"FullClient", "client", "alice", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			user := entity.UserFixtures.Pointer(tc.user)

			var sess *entity.Session

			if tc.role != "" {
				sess = entity.NewSession(3600, 0).SetClient(&entity.Client{ClientRole: tc.role, AuthProvider: "client", AuthScope: "webdav"})
				sess.SetUser(user)
			}

			setWebDAVUser(c, user, sess)

			assert.Equal(t, tc.allowed, canWriteManagedFiles(c.Request.Context()))
			stored, ok := c.Get(gin.AuthUserKey)
			assert.True(t, ok)
			assert.Same(t, user, stored)
		})
	}
}

// TestSetWebDAVUserFullAccess requires full authority even when whole-library reactions are granted.
func TestSetWebDAVUserFullAccess(t *testing.T) {
	for _, role := range []acl.Role{acl.RoleAdmin, acl.RoleClient} {
		t.Run(role.String(), func(t *testing.T) {
			previous := acl.Rules[acl.ResourcePhotos]
			grants := make(acl.Roles, len(previous))
			for key, grant := range previous {
				grants[key] = grant
			}
			grants[role] = acl.Grant{acl.AccessAll: true, acl.ActionReact: true}
			acl.Rules[acl.ResourcePhotos] = grants
			t.Cleanup(func() { acl.Rules[acl.ResourcePhotos] = previous })
			user := entity.UserFixtures.Pointer("alice")
			sess := entity.NewSession(3600, 0).SetClient(&entity.Client{ClientRole: "client", AuthProvider: "client", AuthScope: "webdav"}).SetUser(user)
			assert.True(t, sess.Grants(acl.ResourcePhotos, acl.AccessAll))
			assert.True(t, sess.Grants(acl.ResourcePhotos, acl.ActionReact))
			assert.False(t, sess.Grants(acl.ResourcePhotos, acl.FullAccess))
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPut, "/originals/photo.jpg", nil)
			setWebDAVUser(c, user, sess)
			assert.False(t, canWriteManagedFiles(c.Request.Context()))
			if role == acl.RoleAdmin {
				setWebDAVUser(c, user, nil)
				assert.False(t, canWriteManagedFiles(c.Request.Context()))
			}
		})
	}
}

// TestWebDAVAuthSession_RemovedRow checks that a session whose row was removed is refused and stays removed.
func TestWebDAVAuthSession_RemovedRow(t *testing.T) {
	alice := entity.UserFixtures.Pointer("alice")

	s := entity.NewSession(3600, 0).SetUser(alice).SetScope("webdav")
	s.SetClientName("webdav-removed")
	s.SetClientIP("10.1.1.1")
	s.SetUserAgent("agent-a")
	assert.NoError(t, s.Save())

	token := s.AuthToken()

	// Load the session into the cache, then remove the row directly.
	_, err := entity.FindSession(s.ID)
	assert.NoError(t, err)
	assert.NoError(t, entity.UnscopedDb().Exec("DELETE FROM auth_sessions WHERE id = ?", s.ID).Error)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodGet, "/originals/", nil)
	c.Request.Header.Set("User-Agent", "agent-b")
	c.Request.RemoteAddr = "10.2.2.2:1234"

	sess, user, _, _ := WebDAVAuthSession(c, token)

	assert.Nil(t, sess)
	assert.Nil(t, user)

	var n int
	assert.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("id = ?", s.ID).Count(&n).Error)
	assert.Equal(t, 0, n, "the session row must stay deleted")
}
