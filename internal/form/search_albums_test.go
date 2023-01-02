package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlbumSearchForm(t *testing.T) {
	form := &SearchAlbums{}

	assert.IsType(t, new(SearchAlbums), form)
}

func TestParseQueryStringAlbum(t *testing.T) {
	t.Run("valid query", func(t *testing.T) {
		form := &SearchAlbums{Query: "slug:album1 favorite:true", Year: "2020"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "album1", form.Slug)
		assert.Equal(t, true, form.Favorite)
		assert.Equal(t, "2020", form.Year)
		assert.Equal(t, 0, form.Count)
	})
	t.Run("valid query 2", func(t *testing.T) {
		form := &SearchAlbums{Query: "title:album1 favorite:false q:\"query text\""}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "album1", form.Title)
		assert.Equal(t, false, form.Favorite)
		assert.Equal(t, "", form.Year)
		assert.Equal(t, 0, form.Offset)
		assert.Equal(t, "", form.Order)
		assert.Equal(t, "query text", form.Query)
	})
	t.Run("valid query with umlauts", func(t *testing.T) {
		form := &SearchAlbums{Query: "q:\"tübingen\" year:1999"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "tübingen", form.Query)
		assert.Equal(t, "1999", form.Year)
	})
	t.Run("query for invalid filter", func(t *testing.T) {
		form := &SearchAlbums{Query: "xxx:false"}

		err := form.ParseQueryString()

		if err == nil {
			t.FailNow()
		}

		// log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: xxx", err.Error())
	})
	t.Run("query for favorites with uncommon bool value", func(t *testing.T) {
		form := &SearchAlbums{Query: "favorite:cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.True(t, form.Favorite)
	})
	t.Run("query for count with invalid type", func(t *testing.T) {
		form := &SearchAlbums{Query: "count:cat"}

		err := form.ParseQueryString()

		assert.Error(t, err)
	})
}

func TestNewAlbumSearch(t *testing.T) {
	r := NewAlbumSearch("holiday")
	assert.IsType(t, SearchAlbums{}, r)
}
