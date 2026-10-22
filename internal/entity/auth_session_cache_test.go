package entity

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSessionNotFound(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		assert.True(t, SessionNotFound(ErrSessionIdInvalid))
		assert.True(t, SessionNotFound(ErrSessionNotFound))
		assert.True(t, SessionNotFound(ErrSessionExpired))
		assert.True(t, SessionNotFound(fmt.Errorf("%w %s", ErrSessionIdInvalid, "'x'")))
	})
	t.Run("Other", func(t *testing.T) {
		assert.False(t, SessionNotFound(nil))
		assert.False(t, SessionNotFound(errors.New("no such table: auth_sessions")))
	})
}

func TestFlushSessionCache(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		CacheWebDAVUser("flush-test", NewUser(), CurrentAuthCacheGeneration())
		require.NotPanics(t, func() { FlushSessionCache() })
		assert.Equal(t, 0, sessionCache.ItemCount())
		assert.Nil(t, CachedWebDAVUser("flush-test"))
	})
}

func TestFindSessionByAuthToken(t *testing.T) {
	t.Run("EmptyID", func(t *testing.T) {
		if _, err := FindSessionByAuthToken(""); err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("InvalidID", func(t *testing.T) {
		if _, err := FindSessionByAuthToken("as6sg6bxpogaaba7"); err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		if _, err := FindSessionByAuthToken(rnd.AuthToken()); err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("Alice", func(t *testing.T) {
		if result, err := FindSessionByAuthToken("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"), result.ID)
			assert.Equal(t, UserFixtures.Pointer("alice").UserUID, result.UserUID)
			assert.Equal(t, UserFixtures.Pointer("alice").UserName, result.UserName)
		}
		if cached, err := FindSessionByAuthToken("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"), cached.ID)
			assert.Equal(t, UserFixtures.Pointer("alice").UserUID, cached.UserUID)
			assert.Equal(t, UserFixtures.Pointer("alice").UserName, cached.UserName)
		}
	})
	t.Run("Bob", func(t *testing.T) {
		if result, err := FindSessionByAuthToken("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"), result.ID)
			assert.Equal(t, UserFixtures.Pointer("bob").UserUID, result.UserUID)
			assert.Equal(t, UserFixtures.Pointer("bob").UserName, result.UserName)
		}
		if cached, err := FindSessionByAuthToken("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"), cached.ID)
			assert.Equal(t, UserFixtures.Pointer("bob").UserUID, cached.UserUID)
			assert.Equal(t, UserFixtures.Pointer("bob").UserName, cached.UserName)
		}
	})
	t.Run("Visitor", func(t *testing.T) {
		if result, err := FindSessionByAuthToken("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3"), result.ID)
			assert.Equal(t, Visitor.UserUID, result.UserUID)
			assert.Equal(t, Visitor.UserName, result.UserName)
		}
		if cached, err := FindSessionByAuthToken("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3"), cached.ID)
			assert.Equal(t, Visitor.UserUID, cached.UserUID)
			assert.Equal(t, Visitor.UserName, cached.UserName)
		}
	})
}

func TestFindSession(t *testing.T) {
	t.Run("EmptyID", func(t *testing.T) {
		_, err := FindSession("")
		assert.ErrorIs(t, err, ErrSessionIdInvalid)
	})
	t.Run("InvalidID", func(t *testing.T) {
		_, err := FindSession("as6sg6bxpogaaba7")
		assert.ErrorIs(t, err, ErrSessionIdInvalid)
	})
	t.Run("NotFound", func(t *testing.T) {
		_, err := FindSession(rnd.SessionID(rnd.AuthToken()))
		assert.ErrorIs(t, err, ErrSessionNotFound)
		assert.True(t, SessionNotFound(err))
	})
	t.Run("DatabaseError", func(t *testing.T) {
		originalProvider := dbConn
		tempConn := &DbConn{Driver: dsn.DriverSQLite3, Dsn: fmt.Sprintf("%s/%s", t.TempDir(), "find-session-error.db")}

		SetDbProvider(tempConn)
		t.Cleanup(func() {
			SetDbProvider(originalProvider)
			tempConn.Close()
		})

		_, err := FindSession(rnd.SessionID(rnd.AuthToken()))
		require.Error(t, err)
		assert.False(t, SessionNotFound(err))
	})
	t.Run("Alice", func(t *testing.T) {
		if result, err := FindSession(rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"), result.ID)
			assert.Equal(t, UserFixtures.Pointer("alice").UserUID, result.UserUID)
			assert.Equal(t, UserFixtures.Pointer("alice").UserName, result.UserName)
		}
		if cached, err := FindSession(rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"), cached.ID)
			assert.Equal(t, UserFixtures.Pointer("alice").UserUID, cached.UserUID)
			assert.Equal(t, UserFixtures.Pointer("alice").UserName, cached.UserName)
		}
	})
	t.Run("Bob", func(t *testing.T) {
		if result, err := FindSession(rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"), result.ID)
			assert.Equal(t, UserFixtures.Pointer("bob").UserUID, result.UserUID)
			assert.Equal(t, UserFixtures.Pointer("bob").UserName, result.UserName)
		}
		if cached, err := FindSession(rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"), cached.ID)
			assert.Equal(t, UserFixtures.Pointer("bob").UserUID, cached.UserUID)
			assert.Equal(t, UserFixtures.Pointer("bob").UserName, cached.UserName)
		}
	})
	t.Run("Visitor", func(t *testing.T) {
		if result, err := FindSession(rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3"), result.ID)
			assert.Equal(t, Visitor.UserUID, result.UserUID)
			assert.Equal(t, Visitor.UserName, result.UserName)
		}
		if cached, err := FindSession(rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3"), cached.ID)
			assert.Equal(t, Visitor.UserUID, cached.UserUID)
			assert.Equal(t, Visitor.UserName, cached.UserName)
		}
	})
}

func TestCacheSession(t *testing.T) {
	t.Run("Bob", func(t *testing.T) {
		sessionCache.Flush()
		r, b := sessionCache.Get(rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"))
		assert.Empty(t, r)
		assert.False(t, b)
		bob := FindSessionByRefID("sessxkkcabce")
		CacheSession(bob, time.Hour)
		r2, b2 := sessionCache.Get(rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"))
		assert.NotEmpty(t, r2)
		assert.True(t, b2)
		sessionCache.Flush()
		r3, b3 := sessionCache.Get(rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"))
		assert.Empty(t, r3)
		assert.False(t, b3)
	})
	t.Run("DurationZero", func(t *testing.T) {
		r, b := sessionCache.Get(rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"))
		assert.Empty(t, r)
		assert.False(t, b)
		alice := FindSessionByRefID("sessxkkcabcd")
		CacheSession(alice, 0)
		r2, b2 := sessionCache.Get(rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"))
		assert.NotEmpty(t, r2)
		assert.True(t, b2)
		sessionCache.Flush()
	})
	t.Run("InvalidId", func(t *testing.T) {
		r, b := sessionCache.Get("xxx")
		assert.Empty(t, r)
		assert.False(t, b)
		m := &Session{ID: "xxx"}
		CacheSession(m, 0)
		r2, b2 := sessionCache.Get("xxx")
		assert.Empty(t, r2)
		assert.False(t, b2)
		sessionCache.Flush()
	})
}

func TestFindAnySession_ResolvesClient(t *testing.T) {
	t.Run("ClientSession", func(t *testing.T) {
		FlushSessionCache()
		t.Cleanup(FlushSessionCache)

		fixture := SessionFixtures.Get("client_metrics")
		found, err := FindAnySession(fixture.ID)
		require.NoError(t, err)

		// The client is resolved before the session is cached, so readers of the cached session only read it.
		require.NotNil(t, found.client)
		assert.Equal(t, fixture.ClientUID, found.client.ClientUID)
		_, cached := sessionCache.Get(fixture.ID)
		assert.True(t, cached)

		var wg sync.WaitGroup

		for range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.Equal(t, acl.RoleClient, found.GetClientRole())
				assert.NotEmpty(t, found.Scope())
				assert.NotNil(t, found.GetClient())
			}()
		}

		wg.Wait()
	})
	t.Run("ClientLookupError", func(t *testing.T) {
		FlushSessionCache()
		FlushClientCache()
		t.Cleanup(FlushSessionCache)

		// A session whose client cannot be read is returned, but not cached.
		name := "test:fail_client_reads_" + rnd.Base36(8)
		require.NoError(t, Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
			if db.Statement.Table == (Client{}).TableName() {
				_ = db.AddError(errors.New("client lookup failed"))
			}
		}))
		t.Cleanup(func() { _ = Db().Callback().Query().Remove(name) })

		fixture := SessionFixtures.Get("client_metrics")
		found, err := FindAnySession(fixture.ID)
		require.NoError(t, err)
		assert.Nil(t, found.client)
		_, cached := sessionCache.Get(fixture.ID)
		assert.False(t, cached)
	})
	t.Run("ClientNotFound", func(t *testing.T) {
		FlushSessionCache()
		t.Cleanup(FlushSessionCache)

		// A session of a client that no longer exists is cached without a client.
		sess := NewSession(3600, 0)
		sess.ClientUID = rnd.GenerateUID(ClientUID)
		require.NoError(t, sess.Create())
		t.Cleanup(func() { _ = sess.Delete() })
		sessionCache.Delete(sess.ID)

		found, err := FindAnySession(sess.ID)
		require.NoError(t, err)
		assert.Nil(t, found.client)
		_, cached := sessionCache.Get(sess.ID)
		assert.True(t, cached)
	})
	t.Run("OwnerLookupError", func(t *testing.T) {
		FlushSessionCache()
		FlushClientCache()
		t.Cleanup(FlushSessionCache)

		// A session whose client owner cannot be read is returned, but not cached.
		fixture := ClientFixtures.Get("alice")
		sess := NewSession(3600, 0)
		sess.ClientUID = fixture.ClientUID
		sess.UserUID = fixture.UserUID
		require.NoError(t, sess.Create())
		t.Cleanup(func() { _ = sess.Delete() })
		sessionCache.Delete(sess.ID)

		name := "test:fail_owner_reads_" + rnd.Base36(8)
		require.NoError(t, Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
			if db.Statement.Table == (User{}).TableName() {
				_ = db.AddError(errors.New("user lookup failed"))
			}
		}))
		t.Cleanup(func() { _ = Db().Callback().Query().Remove(name) })

		found, err := FindAnySession(sess.ID)
		require.NoError(t, err)
		assert.Equal(t, fixture.UserUID, found.UserUID)
		_, cached := sessionCache.Get(sess.ID)
		assert.False(t, cached)
	})
	t.Run("UserSession", func(t *testing.T) {
		FlushSessionCache()
		t.Cleanup(FlushSessionCache)

		found, err := FindAnySession(SessionFixtures.Get("alice").ID)
		require.NoError(t, err)
		assert.Nil(t, found.client)
	})
}

func TestSession_setClient(t *testing.T) {
	client := ClientFixtures.Get("metrics")

	t.Run("KeepAuth", func(t *testing.T) {
		m := (&Session{AuthProvider: "access_token", AuthMethod: "session"}).setClient(&client, true)
		assert.Equal(t, "access_token", m.AuthProvider)
		assert.Equal(t, "session", m.AuthMethod)
		assert.Equal(t, client.ClientUID, m.ClientUID)
		assert.Equal(t, client.Scope(), m.AuthScope)
	})
	t.Run("SetAuth", func(t *testing.T) {
		m := (&Session{AuthProvider: "access_token", AuthMethod: "session"}).setClient(&client, false)
		assert.Equal(t, client.Provider().String(), m.AuthProvider)
		assert.Equal(t, client.Method().String(), m.AuthMethod)
	})
	t.Run("NoClient", func(t *testing.T) {
		m := (&Session{AuthProvider: "access_token"}).setClient(nil, false)
		assert.Equal(t, "access_token", m.AuthProvider)
		assert.Nil(t, m.client)
	})
}

func TestSession_SetClientOwnerUnavailable(t *testing.T) {
	FlushClientCache()
	name := "test:fail_user_reads_" + rnd.Base36(8)
	require.NoError(t, Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
		if db.Statement.Table == (User{}).TableName() {
			_ = db.AddError(errors.New("user lookup failed"))
		}
	}))
	t.Cleanup(func() { _ = Db().Callback().Query().Remove(name) })

	// The session keeps the owner of the client when the owner's account cannot be read.
	fixture := ClientFixtures.Get("alice")
	client := Client{ClientUID: fixture.ClientUID, UserUID: fixture.UserUID, UserName: fixture.UserName}
	require.NotEmpty(t, client.UserUID)
	m := (&Session{}).SetClient(&client)
	assert.Equal(t, client.UserUID, m.UserUID)
	assert.False(t, m.NoUser())
	assert.Nil(t, m.user)
}

func TestSession_SetClientKeepsResolvedOwner(t *testing.T) {
	FlushClientCache()
	fixture := ClientFixtures.Get("alice")
	owner := FindUserByUID(fixture.UserUID)
	require.NotNil(t, owner)
	owner.UserScope = "photos"

	m := &Session{ClientUID: fixture.ClientUID, AuthScope: "*"}
	m.SetUser(owner)
	require.Equal(t, "photos", m.Scope())

	name := "test:fail_user_reads_" + rnd.Base36(8)
	require.NoError(t, Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
		if db.Statement.Table == (User{}).TableName() {
			_ = db.AddError(errors.New("user lookup failed"))
		}
	}))
	t.Cleanup(func() { _ = Db().Callback().Query().Remove(name) })

	// Assigning the client keeps the owner already resolved, and with it the owner's scope.
	client := Client{ClientUID: fixture.ClientUID, UserUID: fixture.UserUID, UserName: fixture.UserName, AuthScope: "*"}
	m.setClient(&client, true)
	assert.Same(t, owner, m.user)
	assert.Equal(t, "photos", m.Scope())
}

func TestSession_ResolveClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := &Session{ClientUID: ClientFixtures.Get("metrics").ClientUID}
		assert.NoError(t, m.resolveClient())
		require.NotNil(t, m.client)
		assert.Equal(t, m.ClientUID, m.client.ClientUID)
	})
	t.Run("UnknownClient", func(t *testing.T) {
		m := &Session{ClientUID: rnd.GenerateUID(ClientUID)}
		assert.ErrorIs(t, m.resolveClient(), ErrClientNotFound)
		assert.Nil(t, m.client)
	})
	t.Run("KeepsProvider", func(t *testing.T) {
		client := ClientFixtures.Get("metrics")
		m := &Session{ClientUID: client.ClientUID, AuthProvider: "access_token", AuthMethod: "session", AuthScope: "*"}
		assert.NoError(t, m.resolveClient())
		require.NotNil(t, m.client)
		assert.Equal(t, "access_token", m.AuthProvider)
		assert.Equal(t, "session", m.AuthMethod)
		assert.Equal(t, client.Scope(), m.AuthScope)
	})
	t.Run("NoClient", func(t *testing.T) {
		m := &Session{}
		assert.NoError(t, m.resolveClient())
		assert.Nil(t, m.client)
	})
}
