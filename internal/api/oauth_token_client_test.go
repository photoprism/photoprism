package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newTokenTestClient creates an API client with the given scope and a secret, and removes both after the test.
func newTokenTestClient(t *testing.T, scope string) (*entity.Client, string) {
	t.Helper()

	client := entity.NewClient().SetName("Token Check " + rnd.Base36(6)).SetRole(acl.RoleClient.String())
	client.AuthScope = scope
	client.AuthTokens = 20
	require.NoError(t, client.Create())

	secret, err := client.NewSecret()
	require.NoError(t, err)

	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.Session{}, "client_uid = ?", client.ClientUID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Password{}, "uid = ?", client.ClientUID)
		entity.UnscopedDb().Unscoped().Delete(client)
	})

	return client, secret
}

// clientTokenRequest requests an access token with the client credentials grant.
func clientTokenRequest(app http.Handler, clientUID, secret string) *httptest.ResponseRecorder {
	data := url.Values{
		"grant_type":    {authn.GrantClientCredentials.String()},
		"client_id":     {clientUID},
		"client_secret": {secret},
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Add(header.ContentType, header.ContentTypeForm)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	return w
}

// registerMetricsCheck registers a route that answers 200 if the request is authorized to view metrics.
func registerMetricsCheck(router *gin.RouterGroup) {
	router.GET("/metrics-check", func(c *gin.Context) {
		if s := Auth(c, acl.ResourceMetrics, acl.AccessAll); s.Abort(c) {
			return
		}

		c.Status(http.StatusOK)
	})
}

// countClientSessions returns the number of stored sessions of the client.
func countClientSessions(t *testing.T, clientUID string) int64 {
	t.Helper()

	var n int64
	require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("client_uid = ?", clientUID).Count(&n).Error)

	return n
}

// onClientSessionCreate runs change in the transaction that inserts the next session of the client.
func onClientSessionCreate(t *testing.T, clientUID string, change func(tx *gorm.DB) error) {
	t.Helper()

	name := "test:client_session_" + rnd.Base36(8)
	var once sync.Once

	err := entity.Db().Callback().Create().Before("gorm:create").Register(name, func(tx *gorm.DB) {
		if s, ok := tx.Statement.Model.(*entity.Session); !ok || s.ClientUID != clientUID {
			return
		}

		once.Do(func() {
			if changeErr := change(tx.Session(&gorm.Session{NewDB: true})); changeErr != nil {
				_ = tx.AddError(changeErr)
			}
		})
	})

	require.NoError(t, err)

	t.Cleanup(func() {
		_ = entity.Db().Callback().Create().Remove(name)
	})
}

// TestOAuthToken_ClientChanged covers client credentials requests whose stored client differs from the
// authenticated copy when the new session is saved.
func TestOAuthToken_ClientChanged(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	OAuthToken(router)
	registerMetricsCheck(router)

	updateClient := func(clientUID string, values entity.Values) func(tx *gorm.DB) error {
		return func(tx *gorm.DB) error {
			return tx.Table(entity.Client{}.TableName()).Where("client_uid = ?", clientUID).UpdateColumns(values).Error
		}
	}

	cases := []struct {
		name   string
		change func(client *entity.Client) func(tx *gorm.DB) error
	}{
		{"Disabled", func(client *entity.Client) func(tx *gorm.DB) error {
			return updateClient(client.ClientUID, entity.Values{"auth_enabled": false})
		}},
		{"Deleted", func(client *entity.Client) func(tx *gorm.DB) error {
			return updateClient(client.ClientUID, entity.Values{"deleted_at": time.Now().UTC()})
		}},
		{"RoleChanged", func(client *entity.Client) func(tx *gorm.DB) error {
			return updateClient(client.ClientUID, entity.Values{"client_role": acl.RoleNone.String()})
		}},
		{"ScopeNarrowed", func(client *entity.Client) func(tx *gorm.DB) error {
			return updateClient(client.ClientUID, entity.Values{"auth_scope": "vision"})
		}},
		{"SecretChanged", func(client *entity.Client) func(tx *gorm.DB) error {
			pw := entity.NewPassword(client.ClientUID, rnd.ClientSecret(), false)

			return func(tx *gorm.DB) error {
				return tx.Table(entity.Password{}.TableName()).Where("uid = ?", client.ClientUID).UpdateColumn("hash", pw.Hash).Error
			}
		}},
		{"Purged", func(client *entity.Client) func(tx *gorm.DB) error {
			return func(tx *gorm.DB) error {
				return tx.Exec("DELETE FROM auth_clients WHERE client_uid = ?", client.ClientUID).Error
			}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, secret := newTokenTestClient(t, "metrics")
			onClientSessionCreate(t, client.ClientUID, tc.change(client))

			w := clientTokenRequest(app, client.ClientUID, secret)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "body=%s", w.Body.String())
			assert.Empty(t, gjson.Get(w.Body.String(), "access_token").String())
			assert.Equal(t, int64(0), countClientSessions(t, client.ClientUID))
		})
	}

	t.Run("Unchanged", func(t *testing.T) {
		client, secret := newTokenTestClient(t, "metrics")
		onClientSessionCreate(t, client.ClientUID, updateClient(client.ClientUID, entity.Values{"last_active": 1}))

		w := clientTokenRequest(app, client.ClientUID, secret)
		require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())

		token := gjson.Get(w.Body.String(), "access_token").String()
		assert.Equal(t, http.StatusOK, AuthenticatedRequest(app, http.MethodGet, "/api/v1/metrics-check", token).Code)
	})
	t.Run("UnchangedUserClient", func(t *testing.T) {
		client, secret := newTokenTestClient(t, "metrics")
		user := entity.UserFixtures.Pointer("alice")
		require.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).Where("client_uid = ?", client.ClientUID).
			UpdateColumns(entity.Values{"user_uid": user.UserUID, "user_name": user.UserName}).Error)

		w := clientTokenRequest(app, client.ClientUID, secret)
		require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())

		token := gjson.Get(w.Body.String(), "access_token").String()
		assert.Equal(t, http.StatusOK, AuthenticatedRequest(app, http.MethodGet, "/api/v1/metrics-check", token).Code)
	})
	t.Run("UnchangedNodeClient", func(t *testing.T) {
		client, secret := newTokenTestClient(t, "metrics")
		require.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).Where("client_uid = ?", client.ClientUID).
			UpdateColumns(entity.Values{"node_uuid": rnd.UUIDv7(), "client_role": acl.RoleInstance.String()}).Error)

		w := clientTokenRequest(app, client.ClientUID, secret)
		require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
		assert.Equal(t, int64(1), countClientSessions(t, client.ClientUID))
	})
	t.Run("EmptyScope", func(t *testing.T) {
		client, secret := newTokenTestClient(t, "")
		require.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).Where("client_uid = ?", client.ClientUID).UpdateColumn("auth_scope", "").Error)

		w := clientTokenRequest(app, client.ClientUID, secret)
		require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
		assert.Equal(t, int64(1), countClientSessions(t, client.ClientUID))
	})
	t.Run("ParallelRequests", func(t *testing.T) {
		client, secret := newTokenTestClient(t, "metrics")

		const n = 8
		tokens := make([]string, n)
		codes := make([]int, n)

		var wg sync.WaitGroup

		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				w := clientTokenRequest(app, client.ClientUID, secret)
				codes[i] = w.Code
				tokens[i] = gjson.Get(w.Body.String(), "access_token").String()
			}(i)
		}

		wg.Wait()

		for i := 0; i < n; i++ {
			if assert.Equal(t, http.StatusOK, codes[i]) {
				assert.Equal(t, http.StatusOK, AuthenticatedRequest(app, http.MethodGet, "/api/v1/metrics-check", tokens[i]).Code)
			}
		}
	})
}

// TestAuthorizeSession_ClientDisabled covers sessions of a disabled client, read from the database and from
// the session cache.
func TestAuthorizeSession_ClientDisabled(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	OAuthToken(router)
	registerMetricsCheck(router)

	newToken := func(t *testing.T) (*entity.Client, string) {
		t.Helper()

		client, secret := newTokenTestClient(t, "metrics")
		w := clientTokenRequest(app, client.ClientUID, secret)
		require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())

		token := gjson.Get(w.Body.String(), "access_token").String()
		require.Equal(t, http.StatusOK, AuthenticatedRequest(app, http.MethodGet, "/api/v1/metrics-check", token).Code)

		return client, token
	}

	t.Run("Cached", func(t *testing.T) {
		client, token := newToken(t)
		require.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).Where("client_uid = ?", client.ClientUID).UpdateColumn("auth_enabled", false).Error)

		assert.Equal(t, http.StatusUnauthorized, AuthenticatedRequest(app, http.MethodGet, "/api/v1/metrics-check", token).Code)
	})
	t.Run("Stored", func(t *testing.T) {
		client, token := newToken(t)
		require.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).Where("client_uid = ?", client.ClientUID).UpdateColumn("auth_enabled", false).Error)
		entity.FlushSessionCache()

		assert.Equal(t, http.StatusUnauthorized, AuthenticatedRequest(app, http.MethodGet, "/api/v1/metrics-check", token).Code)
	})
	t.Run("Deleted", func(t *testing.T) {
		client, token := newToken(t)
		require.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).Where("client_uid = ?", client.ClientUID).UpdateColumn("deleted_at", time.Now().UTC()).Error)

		assert.Equal(t, http.StatusUnauthorized, AuthenticatedRequest(app, http.MethodGet, "/api/v1/metrics-check", token).Code)
	})
	t.Run("Enabled", func(t *testing.T) {
		_, token := newToken(t)

		assert.Equal(t, http.StatusOK, AuthenticatedRequest(app, http.MethodGet, "/api/v1/metrics-check", token).Code)
	})
}

func TestRevokeSession(t *testing.T) {
	t.Run("Enabled", func(t *testing.T) {
		_, sess := newOwnedClientSession(t, "*")
		assert.NotNil(t, revokeSession(sess.AuthToken()))
	})
	t.Run("ClientDisabled", func(t *testing.T) {
		_, sess := newOwnedClientSession(t, "*")
		require.NotNil(t, revokeSession(sess.AuthToken()))
		require.NoError(t, disableStoredClient(sess.ClientUID))
		assert.Nil(t, revokeSession(sess.AuthToken()))
	})
	t.Run("InvalidToken", func(t *testing.T) {
		assert.Nil(t, revokeSession("invalid"))
	})
}

// TestOAuthRevoke_ClientDisabled covers a revoke request of a disabled client for its own token.
func TestOAuthRevoke_ClientDisabled(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	OAuthRevoke(router)

	// Leave one token, so a single charge makes Reject return true.
	prevLimit := limiter.Auth
	limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), 3)
	limiter.Auth.ReserveN("198.51.100.7", 2)
	t.Cleanup(func() { limiter.Auth = prevLimit })

	_, sess := newOwnedClientSession(t, "*")
	require.NoError(t, disableStoredClient(sess.ClientUID))

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/revoke", nil)
	req.RemoteAddr = "198.51.100.7:1234"
	req.Header.Add(header.XAuthToken, sess.AuthToken())

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	assert.Equal(t, int64(0), countClientSessions(t, sess.ClientUID))
	assert.False(t, limiter.Auth.Reject("198.51.100.7"))
}
