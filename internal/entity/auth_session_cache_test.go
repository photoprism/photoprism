package entity

import (
	"testing"

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
