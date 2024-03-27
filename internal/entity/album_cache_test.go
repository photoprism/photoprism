package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlushAlbumCache(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		FlushAlbumCache()
	})
}

func TestCachedAlbumByUID(t *testing.T) {
	t.Run("EmptyUID", func(t *testing.T) {
		if _, err := CachedAlbumByUID(""); err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("InvalidUID", func(t *testing.T) {
		if _, err := CachedAlbumByUID("fxgsdrgrg"); err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("as6sg6bxpogaaba7", func(t *testing.T) {
		if result, err := CachedAlbumByUID("as6sg6bxpogaaba7"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "as6sg6bxpogaaba7", result.AlbumUID)
			assert.Equal(t, "christmas-2030", result.AlbumSlug)
		}
		if cached, err := CachedAlbumByUID("as6sg6bxpogaaba7"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "as6sg6bxpogaaba7", cached.AlbumUID)
			assert.Equal(t, "christmas-2030", cached.AlbumSlug)
		}
	})
	t.Run("as6sg6bipotaab23", func(t *testing.T) {
		if result, err := CachedAlbumByUID("as6sg6bipotaab23"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "as6sg6bipotaab23", result.AlbumUID)
			assert.Equal(t, "pest&dogs", result.AlbumSlug)
		}
		if cached, err := CachedAlbumByUID("as6sg6bipotaab23"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "as6sg6bipotaab23", cached.AlbumUID)
			assert.Equal(t, "pest&dogs", cached.AlbumSlug)
		}
	})
}
