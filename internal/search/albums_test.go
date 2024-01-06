package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func TestAlbumPhotos(t *testing.T) {
	t.Run("search with string", func(t *testing.T) {
		results, err := AlbumPhotos(entity.AlbumFixtures.Get("april-1990"), 2, true)

		if err != nil {
			t.Fatal(err)
		}

		if len(results) < 2 {
			t.Errorf("at least 2 results expected: %d", len(results))
		}
	})
}

func TestAlbums(t *testing.T) {
	t.Run("search with string", func(t *testing.T) {
		query := form.NewAlbumSearch("chr")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Christmas 2030", result[0].AlbumTitle)
	})

	t.Run("search with slug", func(t *testing.T) {
		query := form.NewAlbumSearch("slug:holiday")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Holiday 2030", result[0].AlbumTitle)
	})

	t.Run("search with country", func(t *testing.T) {
		query := form.NewAlbumSearch("country:ca")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "April 1990", result[0].AlbumTitle)
	})

	t.Run("favorites true", func(t *testing.T) {
		query := form.NewAlbumSearch("favorite:true")
		query.Count = 100000

		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Holiday 2030", result[0].AlbumTitle)
	})
	t.Run("empty query", func(t *testing.T) {
		query := form.NewAlbumSearch("")

		results, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		if len(results) < 3 {
			t.Errorf("at least 3 results expected: %d", len(results))
		}
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
		f := form.SearchAlbums{
			Query:    "",
			UID:      "as6sg6bxpogaaba7",
			Slug:     "",
			Title:    "",
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
		assert.Equal(t, "christmas-2030", result[0].AlbumSlug)
	})
	t.Run("search with multiple filters", func(t *testing.T) {
		f := form.SearchAlbums{
			Query:    "",
			Type:     "moment",
			Category: "Fun",
			Location: "Favorite Park",
			Title:    "Empty Moment",
			Count:    0,
			Offset:   0,
			Order:    "",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(result))
		assert.Equal(t, "Empty Moment", result[0].AlbumTitle)
	})
	t.Run("search for  year/month/day", func(t *testing.T) {
		f := form.SearchAlbums{
			Year:   "2021",
			Month:  "10",
			Day:    "3",
			Count:  0,
			Offset: 0,
			Order:  "",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(result))
	})
	t.Run("SearchAlbumForYear", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumManual,
			Year:   "2018",
			Month:  "",
			Day:    "",
			Count:  10,
			Offset: 0,
			Order:  "added",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(result))
	})
	t.Run("Folders", func(t *testing.T) {
		query := form.NewAlbumSearch("19")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "April 1990", result[0].AlbumTitle)
	})
	t.Run("California", func(t *testing.T) {
		query := form.NewAlbumSearch("california")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("albums: %#v", result)

		assert.GreaterOrEqual(t, 3, len(result))
	})
	t.Run("Blue", func(t *testing.T) {
		query := form.NewAlbumSearch("blue")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(result))
	})
}
