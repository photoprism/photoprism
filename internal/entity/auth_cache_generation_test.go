package entity

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestAuthCacheGeneration checks targeted and global invalidation of pending credential writes.
func TestAuthCacheGeneration(t *testing.T) {
	user, other := NewUser(), NewUser()
	generation := CurrentAuthCacheGeneration()
	key, secondKey, otherKey := rnd.AuthToken(), rnd.AuthToken(), rnd.AuthToken()
	t.Cleanup(func() {
		FlushUserSessionCache(user.UserUID)
		FlushUserSessionCache(other.UserUID)
	})
	CacheWebDAVUser(key, user, generation)
	CacheWebDAVUser(secondKey, user, generation)
	CacheWebDAVUser(otherKey, other, generation)
	FlushUserSessionCache(user.UserUID)
	for _, k := range []string{key, secondKey} {
		assert.Nil(t, CachedWebDAVUser(k))
		CacheWebDAVUser(k, user, generation)
		assert.Nil(t, CachedWebDAVUser(k))
	}
	assert.Same(t, other, CachedWebDAVUser(otherKey))
	webDAVUserCache.Delete(otherKey)
	CacheWebDAVUser(otherKey, other, generation)
	assert.Same(t, other, CachedWebDAVUser(otherKey))
	current := CurrentAuthCacheGeneration()
	CacheWebDAVUser(key, user, current)
	assert.Same(t, user, CachedWebDAVUser(key))
	FlushSessionCache()
	CacheWebDAVUser(key, user, current)
	CacheWebDAVUser(otherKey, other, generation)
	assert.Nil(t, CachedWebDAVUser(key))
	assert.Nil(t, CachedWebDAVUser(otherKey))
	CacheWebDAVUser(key, user, CurrentAuthCacheGeneration())
	assert.Same(t, user, CachedWebDAVUser(key))
}

// TestCacheSessionGeneration checks that older session objects cannot restore evicted entries.
func TestCacheSessionGeneration(t *testing.T) {
	user, other := NewUser(), NewUser()
	old := NewSession(-1, -1).SetUser(user)
	untouched := NewSession(-1, -1).SetUser(other)
	t.Cleanup(func() {
		FlushUserSessionCache(user.UserUID)
		FlushUserSessionCache(other.UserUID)
	})
	old.Cache()
	untouched.Cache()
	FlushUserSessionCache(user.UserUID)
	old.Cache()
	_, found := sessionCache.Get(old.ID)
	assert.False(t, found)
	assert.Same(t, user, old.GetUser(), "the caller retains its admitted user")
	DeleteFromSessionCache(untouched.ID)
	untouched.Cache()
	cached, found := sessionCache.Get(untouched.ID)
	require.True(t, found)
	assert.Same(t, untouched, cached)
	fresh := NewSession(-1, -1).SetUser(user)
	fresh.Cache()
	_, found = sessionCache.Get(fresh.ID)
	assert.True(t, found)
	untracked := &Session{ID: rnd.SessionID(rnd.AuthToken()), UserUID: user.UserUID, user: user}
	untracked.Cache()
	_, found = sessionCache.Get(untracked.ID)
	assert.False(t, found)
	unloaded := &Session{ID: rnd.SessionID(rnd.AuthToken()), UserUID: user.UserUID}
	unloaded.Cache()
	_, found = sessionCache.Get(unloaded.ID)
	assert.True(t, found)
	FlushSessionCache()
	fresh.Cache()
	unloaded.Cache()
	_, found = sessionCache.Get(fresh.ID)
	assert.False(t, found)
	_, found = sessionCache.Get(unloaded.ID)
	assert.False(t, found)
}

// TestFindSessionGeneration checks a database lookup that completes after account invalidation.
func TestFindSessionGeneration(t *testing.T) {
	user := createScopedTestUser(t)
	sess, err := AddClientSession("generation", 3600, "*", authn.GrantClientCredentials, user)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sess.Delete()) })
	sess.ClearCache()
	loaded, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var once sync.Once
	var blocked atomic.Bool
	var result *Session
	var findErr error
	callback := "test:find_session_generation"
	Db().Callback().Query().After("gorm:query").Register(callback, func(scope *gorm.Scope) {
		var row *Session
		switch value := scope.Value.(type) {
		case *Session:
			row = value
		case **Session:
			row = *value
		}
		if row != nil && row.ID == sess.ID && blocked.CompareAndSwap(false, true) {
			close(loaded)
			<-release
		}
	})
	t.Cleanup(func() {
		once.Do(func() { close(release) })
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("session lookup did not finish")
		}
		Db().Callback().Query().Remove(callback)
	})
	go func() { result, findErr = FindSession(sess.ID); close(done) }()
	select {
	case <-loaded:
	case <-time.After(5 * time.Second):
		t.Fatal("session lookup did not reach the barrier")
	}
	FlushUserSessionCache(user.UserUID)
	once.Do(func() { close(release) })
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("session lookup did not finish")
	}
	require.NoError(t, findErr)
	require.NotNil(t, result)
	_, found := sessionCache.Get(sess.ID)
	assert.False(t, found)
	result.Cache()
	_, found = sessionCache.Get(sess.ID)
	assert.False(t, found)
	fresh, err := FindSession(sess.ID)
	require.NoError(t, err)
	assert.NotSame(t, result, fresh)
	_, found = sessionCache.Get(sess.ID)
	assert.True(t, found)
}
