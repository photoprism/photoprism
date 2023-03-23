package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/acl"
)

func TestUserMap_Get(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		r := UserFixtures.Get("alice")
		assert.Equal(t, "alice", r.UserName)
		assert.Equal(t, "alice", r.Username())
		assert.IsType(t, User{}, r)
	})

	t.Run("Invalid", func(t *testing.T) {
		r := UserFixtures.Get("monstera")
		assert.Equal(t, "", r.UserName)
		assert.Equal(t, "", r.Username())
		assert.IsType(t, User{}, r)
	})
}

func TestUserMap_Pointer(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		r := UserFixtures.Pointer("alice")
		assert.Equal(t, "alice", r.Username())
		assert.Equal(t, "alice", r.UserName)
		assert.Equal(t, "alice@example.com", r.Email())
		assert.Equal(t, "alice@example.com", r.UserEmail)
		assert.Equal(t, acl.RoleAdmin, r.AclRole())
		assert.IsType(t, &User{}, r)
	})

	t.Run("Invalid", func(t *testing.T) {
		r := UserFixtures.Pointer("monstera")
		assert.Equal(t, "", r.UserName)
		assert.Equal(t, "", r.Email())
		assert.Equal(t, acl.RoleUnknown, r.AclRole())
		assert.IsType(t, &User{}, r)
	})
}
