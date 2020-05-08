package form

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAlbum(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var album = struct {
			AlbumName        string
			AlbumDescription string
			AlbumNotes       string
			AlbumOrder       string
			AlbumTemplate    string
			AlbumFavorite    bool
		}{
			AlbumName:        "Foo",
			AlbumDescription: "bar",
			AlbumNotes:       "test notes",
			AlbumOrder:       "newest",
			AlbumTemplate:    "default",
			AlbumFavorite:    true,
		}

		r, err := NewAlbum(album)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Foo", r.AlbumName)
		assert.Equal(t, "bar", r.AlbumDescription)
		assert.Equal(t, "test notes", r.AlbumNotes)
		assert.Equal(t, "newest", r.AlbumOrder)
		assert.Equal(t, "default", r.AlbumTemplate)
		assert.Equal(t, true, r.AlbumFavorite)
	})
}
