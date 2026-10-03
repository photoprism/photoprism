package entity

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
