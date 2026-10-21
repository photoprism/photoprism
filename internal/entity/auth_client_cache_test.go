package entity

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newCacheTestClient stores a client for a cache test and purges it after the test.
func newCacheTestClient(t *testing.T, name string) *Client {
	t.Helper()

	client, err := AddClient(form.Client{ClientName: name, AuthScope: "metrics"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Purge() })

	return client
}

// clientCached reports whether the client cache holds an entry for the UID.
func clientCached(uid string) bool {
	_, hit := clientCache.Get(uid)
	return hit
}

// countClientQueries counts the queries of the auth_clients table until the test ends.
func countClientQueries(t *testing.T) *atomic.Int64 {
	t.Helper()

	name := "test:count_client_queries_" + rnd.Base36(8)
	count := &atomic.Int64{}

	require.NoError(t, Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
		if db.Statement.Table == (Client{}).TableName() {
			count.Add(1)
		}
	}))
	t.Cleanup(func() { _ = Db().Callback().Query().Remove(name) })

	return count
}

func TestCachedClientByUID(t *testing.T) {
	t.Run("InvalidUID", func(t *testing.T) {
		_, err := CachedClientByUID("")
		assert.ErrorIs(t, err, ErrClientNotFound)
		_, err = CachedClientByUID("us5cpu17n6gj2zzz")
		assert.ErrorIs(t, err, ErrClientNotFound)
	})
	t.Run("NotFound", func(t *testing.T) {
		uid := rnd.GenerateUID(ClientUID)
		_, err := CachedClientByUID(uid)
		assert.ErrorIs(t, err, ErrClientNotFound)
		assert.False(t, clientCached(uid))
	})
	t.Run("Hit", func(t *testing.T) {
		client := newCacheTestClient(t, "Client Cache Hit")
		queries := countClientQueries(t)

		for range 5 {
			found, err := CachedClientByUID(client.ClientUID)
			require.NoError(t, err)
			assert.Equal(t, "Client Cache Hit", found.ClientName)
		}

		assert.Equal(t, int64(1), queries.Load())

		// A row removed without a flush is still served from the cache.
		require.NoError(t, UnscopedDb().Delete(&Client{}, "client_uid = ?", client.ClientUID).Error)
		found, err := CachedClientByUID(client.ClientUID)
		require.NoError(t, err)
		assert.Equal(t, client.ClientUID, found.ClientUID)

		FlushClientCache()
		_, err = CachedClientByUID(client.ClientUID)
		assert.ErrorIs(t, err, ErrClientNotFound)
		assert.Equal(t, int64(2), queries.Load())
	})
	t.Run("Error", func(t *testing.T) {
		client := newCacheTestClient(t, "Client Cache Error")
		FlushClientCache()

		name := "test:fail_client_queries_" + rnd.Base36(8)
		require.NoError(t, Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
			if db.Statement.Table == (Client{}).TableName() {
				_ = db.AddError(errors.New("query failed"))
			}
		}))
		removed := false
		t.Cleanup(func() {
			if !removed {
				_ = Db().Callback().Query().Remove(name)
			}
		})

		_, err := CachedClientByUID(client.ClientUID)
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrClientNotFound)
		assert.False(t, clientCached(client.ClientUID))

		require.NoError(t, Db().Callback().Query().Remove(name))
		removed = true

		found, err := CachedClientByUID(client.ClientUID)
		require.NoError(t, err)
		assert.Equal(t, client.ClientUID, found.ClientUID)
	})
	t.Run("Deleted", func(t *testing.T) {
		client := newCacheTestClient(t, "Client Cache Deleted")
		require.NoError(t, client.Delete())

		found, err := CachedClientByUID(client.ClientUID)
		require.NoError(t, err)
		assert.True(t, found.Deleted())
		assert.True(t, found.Disabled())
		assert.True(t, clientCached(client.ClientUID))
	})
	t.Run("Copy", func(t *testing.T) {
		client := newCacheTestClient(t, "Client Cache Copy")
		data := client.GetData()
		data.RedirectURIs = []string{"https://app.example.com/callback"}
		client.SetData(data)
		require.NoError(t, client.SaveChanges("DataJSON"))

		first, err := CachedClientByUID(client.ClientUID)
		require.NoError(t, err)
		require.NotEmpty(t, first.DataJSON)
		assert.Equal(t, []string{"https://app.example.com/callback"}, first.RedirectURIs())

		first.ClientName = "Changed"
		for i := range first.DataJSON {
			first.DataJSON[i] = ' '
		}

		second, err := CachedClientByUID(client.ClientUID)
		require.NoError(t, err)
		assert.Equal(t, "Client Cache Copy", second.ClientName)
		assert.Equal(t, []string{"https://app.example.com/callback"}, second.RedirectURIs())

		// A copy returned from the cache is independent as well.
		second.DataJSON[0] = ' '
		second.ClientName = "Changed"
		third, err := CachedClientByUID(client.ClientUID)
		require.NoError(t, err)
		assert.Equal(t, "Client Cache Copy", third.ClientName)
		assert.Equal(t, []string{"https://app.example.com/callback"}, third.RedirectURIs())
	})
	t.Run("ReadDuringFlush", func(t *testing.T) {
		client := newCacheTestClient(t, "Client Cache Read Flush")
		FlushClientCache()

		// A flush that runs while the row is being read leaves nothing cached.
		name := "test:flush_client_queries_" + rnd.Base36(8)
		require.NoError(t, Db().Callback().Query().After("gorm:query").Register(name, func(db *gorm.DB) {
			if db.Statement.Table == (Client{}).TableName() {
				flushClientCache(client.ClientUID)
			}
		}))
		t.Cleanup(func() { _ = Db().Callback().Query().Remove(name) })

		found, err := CachedClientByUID(client.ClientUID)
		require.NoError(t, err)
		assert.Equal(t, client.ClientUID, found.ClientUID)
		assert.False(t, clientCached(client.ClientUID))
	})
	t.Run("ReadDuringWrite", func(t *testing.T) {
		client := newCacheTestClient(t, "Client Cache Write Read")
		FlushClientCache()

		// A read that runs while the row is being written must not keep the old row cached.
		name := "test:read_client_updates_" + rnd.Base36(8)
		require.NoError(t, Db().Callback().Update().Before("gorm:update").Register(name, func(db *gorm.DB) {
			if db.Statement.Table == (Client{}).TableName() {
				_, _ = CachedClientByUID(client.ClientUID)
			}
		}))
		removed := false
		t.Cleanup(func() {
			if !removed {
				_ = Db().Callback().Update().Remove(name)
			}
		})

		client.ClientName = "Client Cache Written"
		require.NoError(t, client.SaveChanges("ClientName"))
		require.NoError(t, Db().Callback().Update().Remove(name))
		removed = true

		found, err := CachedClientByUID(client.ClientUID)
		require.NoError(t, err)
		assert.Equal(t, "Client Cache Written", found.ClientName)
	})
	t.Run("SessionClient", func(t *testing.T) {
		client := newCacheTestClient(t, "Client Cache Session")
		FlushClientCache()

		_, err := CachedClientByUID(client.ClientUID)
		require.NoError(t, err)
		require.NoError(t, UnscopedDb().Model(&Client{}).Where("client_uid = ?", client.ClientUID).UpdateColumns(Values{"client_name": "Client Cache Session Changed", "client_role": "admin"}).Error)

		// GetClient and the client role of a session read through the cache.
		assert.Equal(t, "Client Cache Session", (&Session{ClientUID: client.ClientUID}).GetClient().ClientName)
		assert.Equal(t, acl.RoleClient, (&Session{ClientUID: client.ClientUID}).GetClientRole())

		FlushClientCache()
		assert.Equal(t, "Client Cache Session Changed", (&Session{ClientUID: client.ClientUID}).GetClient().ClientName)
		assert.Equal(t, acl.RoleAdmin, (&Session{ClientUID: client.ClientUID}).GetClientRole())
	})
	t.Run("SessionLoad", func(t *testing.T) {
		client := newCacheTestClient(t, "Client Cache Session Load")
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", nil)

		ids := make([]string, 3)

		for i := range ids {
			sess := client.NewSession(ctx, authn.GrantClientCredentials)
			require.NoError(t, sess.Save())
			ids[i] = sess.ID
		}

		FlushSessionCache()
		FlushClientCache()
		queries := countClientQueries(t)

		// Loading several sessions of one client reads the client once.
		for _, id := range ids {
			sess, err := FindSession(id)
			require.NoError(t, err)
			assert.Equal(t, acl.RoleClient, sess.GetClientRole())
			assert.Equal(t, client.ClientUID, sess.GetClient().ClientUID)
		}

		assert.Equal(t, int64(1), queries.Load())

		// A role change through the client methods is seen on the next session load.
		client.ClientRole = acl.RoleAdmin.String()
		require.NoError(t, client.SaveChanges("ClientRole"))
		sess := client.NewSession(ctx, authn.GrantClientCredentials)
		require.NoError(t, sess.Save())
		FlushSessionCache()
		found, err := FindSession(sess.ID)
		require.NoError(t, err)
		assert.Equal(t, acl.RoleAdmin, found.GetClientRole())
	})
	t.Run("Expires", func(t *testing.T) {
		prev := ClientCacheDuration
		ClientCacheDuration = 500 * time.Millisecond
		t.Cleanup(func() { ClientCacheDuration = prev })

		client := newCacheTestClient(t, "Client Cache Expires")
		sess := client.NewSession(nil, authn.GrantClientCredentials)
		assert.False(t, sess.ClientDisabled())

		require.NoError(t, UnscopedDb().Model(&Client{}).Where("client_uid = ?", client.ClientUID).UpdateColumn("auth_enabled", false).Error)
		assert.False(t, sess.ClientDisabled())

		time.Sleep(time.Second)
		assert.True(t, sess.ClientDisabled())
	})
}

func TestClientCacheGeneration(t *testing.T) {
	t.Run("FlushedClient", func(t *testing.T) {
		client := newCacheTestClient(t, "Client Cache Flushed")
		FlushClientCache()

		version := clientCacheVersion()
		stale := Client{ClientUID: client.ClientUID, ClientName: "Stale", AuthEnabled: true}
		flushClientCache(client.ClientUID)
		cacheClient(stale, version)
		assert.False(t, clientCached(client.ClientUID))

		cacheClient(stale, clientCacheVersion())
		assert.True(t, clientCached(client.ClientUID))
	})
	t.Run("FlushedAll", func(t *testing.T) {
		uid := rnd.GenerateUID(ClientUID)
		version := clientCacheVersion()
		FlushClientCache()
		cacheClient(Client{ClientUID: uid}, version)
		assert.False(t, clientCached(uid))
	})
	t.Run("OtherClient", func(t *testing.T) {
		uid := rnd.GenerateUID(ClientUID)
		version := clientCacheVersion()
		flushClientCache(rnd.GenerateUID(ClientUID))
		cacheClient(Client{ClientUID: uid}, version)
		assert.True(t, clientCached(uid))
		FlushClientCache()
	})
	t.Run("ReadBeforeWrite", func(t *testing.T) {
		client := newCacheTestClient(t, "Client Cache Write")
		FlushClientCache()

		// A read that started before a write must not publish the row it read.
		version := clientCacheVersion()
		stale := Client{}
		require.NoError(t, UnscopedDb().First(&stale, "client_uid = ?", client.ClientUID).Error)

		client.AuthEnabled = false
		require.NoError(t, client.SaveChanges("AuthEnabled"))
		cacheClient(stale.cacheCopy(), version)
		assert.False(t, clientCached(client.ClientUID))

		found, err := CachedClientByUID(client.ClientUID)
		require.NoError(t, err)
		assert.False(t, found.AuthEnabled)
	})
}

func TestClient_cacheCopy(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := Client{ClientUID: rnd.GenerateUID(ClientUID), DataJSON: []byte(`{"redirectUris":["https://app.example.com/cb"]}`), secretHash: "hash"}
		m.user = &User{UserUID: "uqxetse3cy5eo9z2"}
		m.data = &ClientData{}

		c := m.cacheCopy()
		assert.Nil(t, c.user)
		assert.Nil(t, c.data)
		assert.Empty(t, c.secretHash)
		assert.Equal(t, m.DataJSON, c.DataJSON)

		c.DataJSON[0] = ' '
		assert.Equal(t, byte('{'), m.DataJSON[0])
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, Client{}.cacheCopy().DataJSON)
	})
}

func TestFlushClientCache(t *testing.T) {
	// primed caches the client and asserts that it is cached.
	primed := func(t *testing.T, uid string) {
		t.Helper()
		_, err := CachedClientByUID(uid)
		require.NoError(t, err)
		require.True(t, clientCached(uid))
	}

	t.Run("Create", func(t *testing.T) {
		m := NewClient()
		m.ClientUID = rnd.GenerateUID(ClientUID)
		m.ClientName = "Client Cache Create"
		cacheClient(Client{ClientUID: m.ClientUID}, clientCacheVersion())
		require.True(t, clientCached(m.ClientUID))
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		require.NoError(t, m.Create())
		assert.False(t, clientCached(m.ClientUID))
	})
	t.Run("Save", func(t *testing.T) {
		m := newCacheTestClient(t, "Client Cache Save")
		primed(t, m.ClientUID)
		m.ClientName = "Client Cache Saved"
		require.NoError(t, m.Save())
		found, err := CachedClientByUID(m.ClientUID)
		require.NoError(t, err)
		assert.Equal(t, "Client Cache Saved", found.ClientName)
	})
	t.Run("SaveChanges", func(t *testing.T) {
		m := newCacheTestClient(t, "Client Cache Changes")
		sess := m.NewSession(nil, authn.GrantClientCredentials)
		assert.False(t, sess.ClientDisabled())
		primed(t, m.ClientUID)

		m.AuthEnabled = false
		require.NoError(t, m.SaveChanges("AuthEnabled"))
		assert.True(t, sess.ClientDisabled())
	})
	t.Run("Updates", func(t *testing.T) {
		m := newCacheTestClient(t, "Client Cache Updates")
		primed(t, m.ClientUID)
		require.NoError(t, m.Updates(Values{"ClientName": "Client Cache Updated"}))
		found, err := CachedClientByUID(m.ClientUID)
		require.NoError(t, err)
		assert.Equal(t, "Client Cache Updated", found.ClientName)
	})
	t.Run("Delete", func(t *testing.T) {
		m := newCacheTestClient(t, "Client Cache Delete")
		sess := m.NewSession(nil, authn.GrantClientCredentials)
		primed(t, m.ClientUID)
		require.NoError(t, m.Delete())
		assert.True(t, sess.ClientDisabled())
	})
	t.Run("Restore", func(t *testing.T) {
		m := newCacheTestClient(t, "Client Cache Restore")
		require.NoError(t, m.Delete())
		primed(t, m.ClientUID)
		require.NoError(t, m.Restore())
		found, err := CachedClientByUID(m.ClientUID)
		require.NoError(t, err)
		assert.False(t, found.Deleted())
	})
	t.Run("Purge", func(t *testing.T) {
		m := newCacheTestClient(t, "Client Cache Purge")
		primed(t, m.ClientUID)
		require.NoError(t, m.Purge())
		_, err := CachedClientByUID(m.ClientUID)
		assert.ErrorIs(t, err, ErrClientNotFound)
	})
	t.Run("NewSecret", func(t *testing.T) {
		m := newCacheTestClient(t, "Client Cache NewSecret")
		primed(t, m.ClientUID)
		_, err := m.NewSecret()
		require.NoError(t, err)
		assert.False(t, clientCached(m.ClientUID))
	})
	t.Run("SetSecret", func(t *testing.T) {
		m := newCacheTestClient(t, "Client Cache SetSecret")
		primed(t, m.ClientUID)
		require.NoError(t, m.SetSecret(rnd.ClientSecret()))
		assert.False(t, clientCached(m.ClientUID))
	})
	t.Run("DeleteSecret", func(t *testing.T) {
		m := newCacheTestClient(t, "Client Cache DeleteSecret")
		require.NoError(t, m.SetSecret(rnd.ClientSecret()))
		primed(t, m.ClientUID)
		require.NoError(t, m.DeleteSecret())
		assert.False(t, clientCached(m.ClientUID))
	})
	t.Run("UpdateLastActive", func(t *testing.T) {
		m := newCacheTestClient(t, "Client Cache LastActive")
		primed(t, m.ClientUID)
		m.UpdateLastActive(true)
		assert.True(t, clientCached(m.ClientUID))
	})
	t.Run("FlushCaches", func(t *testing.T) {
		m := newCacheTestClient(t, "Client Cache FlushCaches")
		primed(t, m.ClientUID)
		FlushCaches()
		assert.False(t, clientCached(m.ClientUID))
	})
}
