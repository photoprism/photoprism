package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserShareMap_Get(t *testing.T) {
	t.Run("AliceAlbum", func(t *testing.T) {
		r := UserShareFixtures.Get("AliceAlbum")
		assert.Equal(t, "The quick brown fox jumps over the lazy dog.", r.Comment)
		assert.Equal(t, "at9lxuqxpogaaba9", r.ShareUID)
		assert.IsType(t, UserShare{}, r)
	})

	t.Run("Invalid", func(t *testing.T) {
		r := UserShareFixtures.Get("monstera")
		assert.Equal(t, "", r.Comment)
		assert.Equal(t, "", r.ShareUID)
		assert.IsType(t, UserShare{}, r)
	})
}

func TestUserShareMap_Pointer(t *testing.T) {
	t.Run("AliceAlbum", func(t *testing.T) {
		r := UserShareFixtures.Pointer("AliceAlbum")
		assert.Equal(t, "The quick brown fox jumps over the lazy dog.", r.Comment)
		assert.Equal(t, "at9lxuqxpogaaba9", r.ShareUID)

		assert.IsType(t, &UserShare{}, r)
	})

	t.Run("Invalid", func(t *testing.T) {
		r := UserShareFixtures.Pointer("monstera")
		assert.Equal(t, "", r.Comment)
		assert.Equal(t, "", r.ShareUID)
		assert.IsType(t, &UserShare{}, r)
	})
}
