package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlbumMap_Get(t *testing.T) {
	t.Run("GetExistingAlbum", func(t *testing.T) {
		r := AlbumFixtures.Get("christmas2030")
		assert.Equal(t, "as6sg6bxpogaaba7", r.AlbumUID)
		assert.Equal(t, "christmas-2030", r.AlbumSlug)
		assert.IsType(t, Album{}, r)
	})
	t.Run("GetNotExistingAlbum", func(t *testing.T) {
		r := AlbumFixtures.Get("Fusion 3333")
		assert.Equal(t, "fusion-3333", r.AlbumSlug)
		assert.IsType(t, Album{}, r)
	})
}

func TestAlbumMap_Pointer(t *testing.T) {
	t.Run("GetExistingAlbumPointer", func(t *testing.T) {
		r := AlbumFixtures.Pointer("christmas2030")
		assert.Equal(t, "as6sg6bxpogaaba7", r.AlbumUID)
		assert.Equal(t, "christmas-2030", r.AlbumSlug)
		assert.IsType(t, &Album{}, r)
	})
	t.Run("GetNotExistingAlbumPointer", func(t *testing.T) {
		r := AlbumFixtures.Pointer("Fusion 444")
		assert.Equal(t, "fusion-444", r.AlbumSlug)
		assert.IsType(t, &Album{}, r)
	})
}
