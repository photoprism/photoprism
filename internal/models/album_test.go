package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAlbum(t *testing.T) {
	t.Run("name Christmas 2018", func(t *testing.T) {
		album := NewAlbum("Christmas 2018")
		assert.Equal(t, "Christmas 2018", album.AlbumName)
		assert.Equal(t, "christmas-2018", album.AlbumSlug)
	})
	t.Run("name empty", func(t *testing.T) {
		album := NewAlbum("")
		assert.Equal(t, "New Album", album.AlbumName)
		assert.Equal(t, "new-album", album.AlbumSlug)
	})
}
