package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlbumMap_Get(t *testing.T) {
	t.Run("get existing album", func(t *testing.T) {
		r := AlbumFixtures.Get("christmas2030")
		assert.Equal(t, "as6sg6bxpogaaba7", r.AlbumUID)
		assert.Equal(t, "christmas-2030", r.AlbumSlug)
		assert.IsType(t, Album{}, r)
	})
	t.Run("get not existing album", func(t *testing.T) {
		r := AlbumFixtures.Get("Fusion 3333")
		assert.Equal(t, "fusion-3333", r.AlbumSlug)
		assert.IsType(t, Album{}, r)
	})
}

func TestAlbumMap_Pointer(t *testing.T) {
	t.Run("get existing album pointer", func(t *testing.T) {
		r := AlbumFixtures.Pointer("christmas2030")
		assert.Equal(t, "as6sg6bxpogaaba7", r.AlbumUID)
		assert.Equal(t, "christmas-2030", r.AlbumSlug)
		assert.IsType(t, &Album{}, r)
	})
	t.Run("get not existing album pointer", func(t *testing.T) {
		r := AlbumFixtures.Pointer("Fusion 444")
		assert.Equal(t, "fusion-444", r.AlbumSlug)
		assert.IsType(t, &Album{}, r)
	})
}
