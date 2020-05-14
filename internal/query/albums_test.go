package query

import (
	"testing"

	form "github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestAlbumByUUID(t *testing.T) {
	t.Run("existing uuid", func(t *testing.T) {
		album, err := AlbumByUUID("at9lxuqxpogaaba7")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Christmas2030", album.AlbumName)
	})

	t.Run("not existing uuid", func(t *testing.T) {
		album, err := AlbumByUUID("3765")
		assert.Error(t, err, "record not found")
		t.Log(album)
	})
}

func TestAlbumThumbByUUID(t *testing.T) {
	t.Run("existing uuid", func(t *testing.T) {
		file, err := AlbumThumbByUUID("at9lxuqxpogaaba8")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("not existing uuid", func(t *testing.T) {
		file, err := AlbumThumbByUUID("3765")
		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestAlbums(t *testing.T) {
	t.Run("search with string", func(t *testing.T) {
		query := form.NewAlbumSearch("chr")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Christmas2030", result[0].AlbumName)
	})

	t.Run("search with slug", func(t *testing.T) {
		query := form.NewAlbumSearch("slug:holiday count:10")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Holiday2030", result[0].AlbumName)
	})

	t.Run("favorites true", func(t *testing.T) {
		query := form.NewAlbumSearch("favorite:true count:10000")

		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Holiday2030", result[0].AlbumName)
	})
	t.Run("empty query", func(t *testing.T) {
		query := form.NewAlbumSearch("order:slug")

		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, len(result))
	})
	t.Run("search with invalid query string", func(t *testing.T) {
		query := form.NewAlbumSearch("xxx:bla")
		result, err := Albums(query)
		assert.Error(t, err, "unknown filter")
		t.Log(result)
	})
	t.Run("search with invalid query string", func(t *testing.T) {
		query := form.NewAlbumSearch("xxx:bla")
		result, err := Albums(query)
		assert.Error(t, err, "unknown filter")
		t.Log(result)
	})
	t.Run("search for existing ID", func(t *testing.T) {
		f := form.AlbumSearch{
			Query:    "",
			ID:       "at9lxuqxpogaaba7",
			Slug:     "",
			Name:     "",
			Favorite: false,
			Count:    0,
			Offset:   0,
			Order:    "",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(result))
		assert.Equal(t, "christmas2030", result[0].AlbumSlug)
	})
}
