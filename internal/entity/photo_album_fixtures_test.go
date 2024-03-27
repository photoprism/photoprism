package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhotoAlbumMap_Get(t *testing.T) {
	t.Run("get existing photoalbum", func(t *testing.T) {
		r := PhotoAlbumFixtures.Get("1", "", "")
		assert.Equal(t, "as6sg6bxpogaaba8", r.AlbumUID)
		assert.Equal(t, "ps6sg6be2lvl0yh7", r.PhotoUID)
		assert.IsType(t, PhotoAlbum{}, r)
	})
	t.Run("get not existing photoalbum", func(t *testing.T) {
		r := PhotoAlbumFixtures.Get("x", "1234", "5678")
		assert.Equal(t, "5678", r.AlbumUID)
		assert.Equal(t, "1234", r.PhotoUID)
		assert.IsType(t, PhotoAlbum{}, r)
	})
}

func TestPhotoAlbumMap_Pointer(t *testing.T) {
	t.Run("get existing photoalbum pointer", func(t *testing.T) {
		r := PhotoAlbumFixtures.Pointer("1", "", "")
		assert.Equal(t, "as6sg6bxpogaaba8", r.AlbumUID)
		assert.Equal(t, "ps6sg6be2lvl0yh7", r.PhotoUID)
		assert.IsType(t, &PhotoAlbum{}, r)
	})
	t.Run("get not existing photoalbum pointer", func(t *testing.T) {
		r := PhotoAlbumFixtures.Pointer("xy", "xxx", "yyy")
		assert.Equal(t, "yyy", r.AlbumUID)
		assert.Equal(t, "xxx", r.PhotoUID)
		assert.IsType(t, &PhotoAlbum{}, r)
	})
}
