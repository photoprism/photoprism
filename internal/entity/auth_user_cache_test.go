package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestCacheWebDAVUser checks successful credential caching and its bounded lifetime.
func TestCacheWebDAVUser(t *testing.T) {
	key := rnd.AuthToken()
	t.Cleanup(func() { webDAVUserCache.Delete(key) })
	assert.Nil(t, CachedWebDAVUser(key))
	CacheWebDAVUser(key, nil, CurrentAuthCacheGeneration())
	assert.Nil(t, CachedWebDAVUser(key))
	user := NewUser()
	CacheWebDAVUser("", user, CurrentAuthCacheGeneration())
	assert.Nil(t, CachedWebDAVUser(""))
	before := time.Now()
	CacheWebDAVUser(key, user, CurrentAuthCacheGeneration())
	assert.Same(t, user, CachedWebDAVUser(key))
	_, expires, found := webDAVUserCache.GetWithExpiration(key)
	require.True(t, found)
	assert.WithinDuration(t, before.Add(time.Minute), expires, time.Second)
}

// TestFlushUserSessionCache checks eviction by account without touching other cached principals.
func TestFlushUserSessionCache(t *testing.T) {
	user, other := NewUser(), NewUser()
	sessions := []*Session{NewSession(-1, -1).SetUser(user), NewSession(-1, -1).SetUser(user), NewSession(-1, -1).SetUser(other), NewSession(-1, -1)}
	for _, sess := range sessions {
		CacheSession(sess, time.Minute)
		t.Cleanup(func() { DeleteFromSessionCache(sess.ID) })
	}
	key, otherKey := rnd.AuthToken(), rnd.AuthToken()
	CacheWebDAVUser(key, user, CurrentAuthCacheGeneration())
	CacheWebDAVUser(otherKey, other, CurrentAuthCacheGeneration())
	t.Cleanup(func() {
		webDAVUserCache.Delete(key)
		webDAVUserCache.Delete(otherKey)
	})

	FlushUserSessionCache("")
	assert.Same(t, user, CachedWebDAVUser(key))
	FlushUserSessionCache(user.UserUID)
	for i, sess := range sessions {
		value, found := sessionCache.Get(sess.ID)
		if i < 2 {
			assert.False(t, found)
		} else {
			assert.True(t, found)
			assert.Same(t, sess, value)
		}
	}
	assert.Nil(t, CachedWebDAVUser(key))
	assert.Same(t, other, CachedWebDAVUser(otherKey))
}

// TestUser_SaveFailureKeepsCache checks that a refused write leaves cached sessions usable.
func TestUser_SaveFailureKeepsCache(t *testing.T) {
	user, other := createScopedTestUser(t), createScopedTestUser(t)
	sess := NewSession(-1, -1).SetUser(user)
	sess.Cache()
	CacheWebDAVUser(sess.ID, user, CurrentAuthCacheGeneration())
	t.Cleanup(func() { FlushUserSessionCache(user.UserUID) })
	fresh := FindUserByUID(user.UserUID)
	require.NotNil(t, fresh)
	fresh.UserUID = other.UserUID
	require.Error(t, fresh.Save())
	cached, found := sessionCache.Get(sess.ID)
	require.True(t, found)
	assert.Same(t, sess, cached)
	assert.Same(t, user, CachedWebDAVUser(sess.ID))
}

// TestUser_SaveRefreshesSessionPaths checks that saved paths reach surviving app credentials.
func TestUser_SaveRefreshesSessionPaths(t *testing.T) {
	for _, field := range []string{"BasePath", "UploadPath"} {
		t.Run(field, func(t *testing.T) {
			user, other := createScopedTestUser(t), createScopedTestUser(t)
			sess, err := AddClientSession("path-cache", 3600, "*", authn.GrantClientCredentials, user)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, sess.Delete()) })
			otherSession, err := AddClientSession("other-cache", 3600, "*", authn.GrantClientCredentials, other)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, otherSession.Delete()) })
			sess.Cache()
			otherSession.Cache()
			CacheWebDAVUser(sess.ID, user, CurrentAuthCacheGeneration())
			CacheWebDAVUser(otherSession.ID, other, CurrentAuthCacheGeneration())
			t.Cleanup(func() { webDAVUserCache.Delete(otherSession.ID) })

			fresh := FindUserByUID(user.UserUID)
			require.NotNil(t, fresh)
			frm, err := fresh.Form()
			require.NoError(t, err)
			if field == "BasePath" {
				frm.BasePath = "mybox"
			} else {
				frm.UploadPath = "mybox"
			}
			require.True(t, fresh.PrivilegeLevelChange(frm))
			require.NoError(t, fresh.SaveForm(frm, fresh, true, true))
			assert.Nil(t, CachedWebDAVUser(sess.ID))
			assert.Same(t, other, CachedWebDAVUser(otherSession.ID))
			reloaded, err := FindSession(sess.ID)
			require.NoError(t, err)
			assert.NotSame(t, sess, reloaded)
			assert.Equal(t, "mybox", reloaded.GetUser().GetUploadPath())
			if field == "BasePath" {
				assert.Equal(t, "mybox", reloaded.GetUser().GetBasePath())
			}
			untouched, err := FindSession(otherSession.ID)
			require.NoError(t, err)
			assert.Same(t, otherSession, untouched)
		})
	}
}
