package query

import (
	"testing"

	form "github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestAlbumByUID(t *testing.T) {
	t.Run("existing uid", func(t *testing.T) {
		album, err := AlbumByUID("at9lxuqxpogaaba7")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Christmas2030", album.AlbumTitle)
	})

	t.Run("not existing uid", func(t *testing.T) {
		album, err := AlbumByUID("3765")
		assert.Error(t, err, "record not found")
		t.Log(album)
	})
}

func TestAlbumThumbByUID(t *testing.T) {
	t.Run("existing uid", func(t *testing.T) {
		file, err := AlbumCoverByUID("at9lxuqxpogaaba8")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("not existing uid", func(t *testing.T) {
		file, err := AlbumCoverByUID("3765")
		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestAlbums(t *testing.T) {
	t.Run("search with string", func(t *testing.T) {
		query := form.NewAlbumSearch("chr")
		result, err := AlbumSearch(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Christmas2030", result[0].AlbumTitle)
	})

	t.Run("search with slug", func(t *testing.T) {
		query := form.NewAlbumSearch("slug:holiday count:10")
		result, err := AlbumSearch(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Holiday2030", result[0].AlbumTitle)
	})

	t.Run("favorites true", func(t *testing.T) {
		query := form.NewAlbumSearch("favorite:true count:10000")

		result, err := AlbumSearch(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Holiday2030", result[0].AlbumTitle)
	})
	t.Run("empty query", func(t *testing.T) {
		query := form.NewAlbumSearch("order:slug")

		result, err := AlbumSearch(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, len(result))
	})
	t.Run("search with invalid query string", func(t *testing.T) {
		query := form.NewAlbumSearch("xxx:bla")
		result, err := AlbumSearch(query)
		assert.Error(t, err, "unknown filter")
		t.Log(result)
	})
	t.Run("search with invalid query string", func(t *testing.T) {
		query := form.NewAlbumSearch("xxx:bla")
		result, err := AlbumSearch(query)
		assert.Error(t, err, "unknown filter")
		t.Log(result)
	})
	t.Run("search for existing ID", func(t *testing.T) {
		f := form.AlbumSearch{
			Query:    "",
			ID:       "at9lxuqxpogaaba7",
			Slug:     "",
			Title:    "",
			Favorite: false,
			Count:    0,
			Offset:   0,
			Order:    "",
		}

		result, err := AlbumSearch(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(result))
		assert.Equal(t, "christmas2030", result[0].AlbumSlug)
	})
}
