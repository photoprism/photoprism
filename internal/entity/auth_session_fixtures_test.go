package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionMap_Get(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		r := SessionFixtures.Get("alice")
		assert.Equal(t, "alice", r.UserName)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", r.AuthToken())
		assert.Equal(t, "a3859489780243a78b331bd44f58255b552dee104041a45c0e79b610f63af2e5", r.ID)
		assert.IsType(t, Session{}, r)
	})

	t.Run("Invalid", func(t *testing.T) {
		r := SessionFixtures.Get("xxx")
		assert.Equal(t, "", r.UserName)
		assert.Equal(t, "", r.ID)
		assert.IsType(t, Session{}, r)
	})
}

func TestSessionMap_Pointer(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		r := SessionFixtures.Pointer("alice")
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", r.AuthToken())
		assert.Equal(t, "a3859489780243a78b331bd44f58255b552dee104041a45c0e79b610f63af2e5", r.ID)
		assert.Equal(t, "alice", r.UserName)
		assert.IsType(t, &Session{}, r)
	})

	t.Run("Invalid", func(t *testing.T) {
		r := SessionFixtures.Pointer("xxx")
		assert.Equal(t, "", r.UserName)
		assert.Equal(t, "", r.ID)
		assert.IsType(t, &Session{}, r)
	})
}
