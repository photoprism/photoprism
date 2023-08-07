package entity

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/rnd"

	"github.com/stretchr/testify/assert"
)

func TestFlushSessionCache(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		FlushSessionCache()
	})
}

func TestFindSession(t *testing.T) {
	t.Run("EmptyID", func(t *testing.T) {
		if _, err := FindSession(""); err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("InvalidID", func(t *testing.T) {
		if _, err := FindSession("at9lxuqxpogaaba7"); err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		if _, err := FindSession(rnd.SessionID()); err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("Alice", func(t *testing.T) {
		if result, err := FindSession("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", result.ID)
			assert.Equal(t, UserFixtures.Pointer("alice").UserUID, result.UserUID)
			assert.Equal(t, UserFixtures.Pointer("alice").UserName, result.UserName)
		}
		if cached, err := FindSession("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", cached.ID)
			assert.Equal(t, UserFixtures.Pointer("alice").UserUID, cached.UserUID)
			assert.Equal(t, UserFixtures.Pointer("alice").UserName, cached.UserName)
		}
	})
	t.Run("Bob", func(t *testing.T) {
		if result, err := FindSession("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1", result.ID)
			assert.Equal(t, UserFixtures.Pointer("bob").UserUID, result.UserUID)
			assert.Equal(t, UserFixtures.Pointer("bob").UserName, result.UserName)
		}
		if cached, err := FindSession("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1", cached.ID)
			assert.Equal(t, UserFixtures.Pointer("bob").UserUID, cached.UserUID)
			assert.Equal(t, UserFixtures.Pointer("bob").UserName, cached.UserName)
		}
	})
	t.Run("Visitor", func(t *testing.T) {
		if result, err := FindSession("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3", result.ID)
			assert.Equal(t, Visitor.UserUID, result.UserUID)
			assert.Equal(t, Visitor.UserName, result.UserName)
		}
		if cached, err := FindSession("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3", cached.ID)
			assert.Equal(t, Visitor.UserUID, cached.UserUID)
			assert.Equal(t, Visitor.UserName, cached.UserName)
		}
	})
}

func TestCacheSession(t *testing.T) {
	t.Run("bob", func(t *testing.T) {
		sessionCache.Flush()
		r, b := sessionCache.Get("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")
		assert.Empty(t, r)
		assert.False(t, b)
		bob := FindSessionByRefID("sessxkkcabce")
		CacheSession(bob, time.Hour)
		r2, b2 := sessionCache.Get("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")
		assert.NotEmpty(t, r2)
		assert.True(t, b2)
		sessionCache.Flush()
		r3, b3 := sessionCache.Get("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")
		assert.Empty(t, r3)
		assert.False(t, b3)
	})
	t.Run("duration 0", func(t *testing.T) {
		r, b := sessionCache.Get("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")
		assert.Empty(t, r)
		assert.False(t, b)
		alice := FindSessionByRefID("sessxkkcabcd")
		CacheSession(alice, 0)
		r2, b2 := sessionCache.Get("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")
		assert.NotEmpty(t, r2)
		assert.True(t, b2)
		sessionCache.Flush()
	})
	t.Run("invalid ID", func(t *testing.T) {
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

func TestDeleteSession(t *testing.T) {
	m := &Session{ID: "77be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1", DownloadToken: "download123", PreviewToken: "preview123"}
	CacheSession(m, time.Hour)
	r, _ := sessionCache.Get("77be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")
	assert.NotEmpty(t, r)
	DeleteSession(m)
	r2, _ := sessionCache.Get("77be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")
	assert.Empty(t, r2)
}

func TestDeleteFromSessionCache(t *testing.T) {
	sessionCache.Flush()
	bob := FindSessionByRefID("sessxkkcabce")
	CacheSession(bob, time.Hour)
	r, b := sessionCache.Get("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")
	assert.NotEmpty(t, r)
	assert.True(t, b)
	DeleteFromSessionCache("")
	r2, b2 := sessionCache.Get("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")
	assert.NotEmpty(t, r2)
	assert.True(t, b2)
	DeleteFromSessionCache("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")
	r3, b3 := sessionCache.Get("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")
	assert.Empty(t, r3)
	assert.False(t, b3)
}
