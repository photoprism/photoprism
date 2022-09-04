package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserMap_Get(t *testing.T) {
	t.Run("get existing user", func(t *testing.T) {
		r := UserFixtures.Get("alice")
		assert.Equal(t, "alice", r.Username)
		assert.Equal(t, "alice", r.UserName())
		assert.IsType(t, User{}, r)
	})
	t.Run("get not existing user", func(t *testing.T) {
		r := UserFixtures.Get("monstera")
		assert.Equal(t, "", r.Username)
		assert.Equal(t, "", r.UserName())
		assert.IsType(t, User{}, r)
	})
}

func TestUserMap_Pointer(t *testing.T) {
	t.Run("get existing user", func(t *testing.T) {
		r := UserFixtures.Pointer("alice")
		assert.Equal(t, "alice", r.Username)
		assert.IsType(t, &User{}, r)
	})
	t.Run("get not existing user", func(t *testing.T) {
		r := UserFixtures.Pointer("monstera")
		assert.Equal(t, "", r.Username)
		assert.IsType(t, &User{}, r)
	})
}
