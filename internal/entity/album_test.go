package entity

import (
	"github.com/photoprism/photoprism/internal/form"
	"testing"
	"time"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/pkg/txt"
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

func TestAlbum_SetName(t *testing.T) {
	t.Run("valid name", func(t *testing.T) {
		album := NewAlbum("initial name")
		assert.Equal(t, "initial name", album.AlbumName)
		assert.Equal(t, "initial-name", album.AlbumSlug)
		album.SetName("New Album Name")
		assert.Equal(t, "New Album Name", album.AlbumName)
		assert.Equal(t, "new-album-name", album.AlbumSlug)
	})
	t.Run("empty name", func(t *testing.T) {
		album := NewAlbum("initial name")
		assert.Equal(t, "initial name", album.AlbumName)
		assert.Equal(t, "initial-name", album.AlbumSlug)

		album.SetName("")
		expected := album.CreatedAt.Format("January 2006")
		assert.Equal(t, expected, album.AlbumName)
		assert.Equal(t, slug.Make(expected), album.AlbumSlug)
	})
	t.Run("long name", func(t *testing.T) {
		longName := `A value in decimal degrees to a precision of 4 decimal places is precise to 11.132 meters at the 
equator. A value in decimal degrees to 5 decimal places is precise to 1.1132 meter at the equator. Elevation also 
introduces a small error. At 6,378 m elevation, the radius and surface distance is increased by 0.001 or 0.1%. 
Because the earth is not flat, the precision of the longitude part of the coordinates increases 
the further from the equator you get. The precision of the latitude part does not increase so much, 
more strictly however, a meridian arc length per 1 second depends on the latitude at the point in question. 
The discrepancy of 1 second meridian arc length between equator and pole is about 0.3 metres because the earth 
is an oblate spheroid.`
		expected := txt.Clip(longName, txt.ClipDefault)
		slugExpected := txt.Clip(longName, txt.ClipSlug)
		album := NewAlbum(longName)
		assert.Equal(t, expected, album.AlbumName)
		assert.Contains(t, album.AlbumSlug, slug.Make(slugExpected))
	})
}

func TestAlbum_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		album := NewAlbum("Old Name")

		assert.Equal(t, "Old Name", album.AlbumName)
		assert.Equal(t, "old-name", album.AlbumSlug)

		album2 := Album{ID: 123, AlbumName: "New name", AlbumDescription: "new description"}

		albumForm, err := form.NewAlbum(album2)

		if err != nil {
			t.Fatal(err)
		}

		err = album.Save(albumForm)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "New name", album.AlbumName)
		assert.Equal(t, "new description", album.AlbumDescription)
	})

}
