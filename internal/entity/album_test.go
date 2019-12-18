package entity

import (
	"testing"
	"time"

	"github.com/gosimple/slug"
	"github.com/stretchr/testify/assert"
)

func TestNewAlbum(t *testing.T) {
	t.Run("name Christmas 2018", func(t *testing.T) {
		album := NewAlbum("Christmas 2018")
		assert.Equal(t, "Christmas 2018", album.AlbumName)
		assert.Equal(t, "christmas-2018", album.AlbumSlug)
	})
	t.Run("name empty", func(t *testing.T) {
		album := NewAlbum("")

		defaultName := time.Now().Format("January 2006")
		defaultSlug := slug.Make(defaultName)

		assert.Equal(t, defaultName, album.AlbumName)
		assert.Equal(t, defaultSlug, album.AlbumSlug)
	})
}

func TestRename(t *testing.T) {
	t.Run("valid name", func(t *testing.T) {
		album := NewAlbum("initial name")
		assert.Equal(t, "initial name", album.AlbumName)
		assert.Equal(t, "initial-name", album.AlbumSlug)
		album.Rename("new album name")
		assert.Equal(t, "new album name", album.AlbumName)
		assert.Equal(t, "new-album-name", album.AlbumSlug)
	})
	t.Run("empty name", func(t *testing.T) {
		album := NewAlbum("initial name")
		assert.Equal(t, "initial name", album.AlbumName)
		assert.Equal(t, "initial-name", album.AlbumSlug)
		t.Log(album.CreatedAt)
		album.Rename("")
		assert.Equal(t, "January 0001", album.AlbumName)
		assert.Equal(t, "january-0001", album.AlbumSlug)
	})
}
