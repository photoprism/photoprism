package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResource_String(t *testing.T) {
	t.Run("Albums", func(t *testing.T) {
		assert.Equal(t, "albums", ResourceAlbums.String())
	})
	t.Run("Favorites", func(t *testing.T) {
		assert.Equal(t, "favorites", ResourceFavorites.String())
	})
	t.Run("Files", func(t *testing.T) {
		assert.Equal(t, "files", ResourceFiles.String())
	})
}

func TestResource_LogId(t *testing.T) {
	t.Run("Albums", func(t *testing.T) {
		assert.Equal(t, "albums", ResourceAlbums.LogId())
	})
	t.Run("Favorites", func(t *testing.T) {
		assert.Equal(t, "favorites", ResourceFavorites.LogId())
	})
	t.Run("Files", func(t *testing.T) {
		assert.Equal(t, "files", ResourceFiles.LogId())
	})
}

func TestResource_Equal(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, ResourceAlbums.Equal("albums"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, ResourceAlbums.Equal("favorites"))
	})
}

func TestResource_NotEqual(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, ResourceAlbums.NotEqual("favorites"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, ResourceAlbums.NotEqual("albums"))
	})
}
