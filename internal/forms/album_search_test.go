package forms

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAlbumSearchForm(t *testing.T) {
	form := &AlbumSearchForm{}

	assert.IsType(t, new(AlbumSearchForm), form)
}

func TestParseQueryStringAlbum(t *testing.T) {

	t.Run("valid query", func(t *testing.T) {
		form := &AlbumSearchForm{Query: "slug:album1 favorites:true count:10"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Nil(t, err)
		assert.Equal(t, "album1", form.Slug)
		assert.Equal(t, true, form.Favorites)
		assert.Equal(t, 10, form.Count)
	})
	t.Run("valid query 2", func(t *testing.T) {
		form := &AlbumSearchForm{Query: "name:album1 favorites:false offset:100 order:newest query:\"query text\""}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Nil(t, err)
		assert.Equal(t, "album1", form.Name)
		assert.Equal(t, false, form.Favorites)
		assert.Equal(t, 100, form.Offset)
		assert.Equal(t, "newest", form.Order)
		assert.Equal(t, "query text", form.Query)
	})
	t.Run("query for invalid filter", func(t *testing.T) {
		form := &AlbumSearchForm{Query: "xxx:false"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: Xxx", err.Error())
	})
	t.Run("query for favorites with invalid type", func(t *testing.T) {
		form := &AlbumSearchForm{Query: "favorites:cat"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Equal(t, "not a bool value: Favorites", err.Error())
	})
	t.Run("query for count with invalid type", func(t *testing.T) {
		form := &AlbumSearchForm{Query: "count:cat"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Equal(t, "strconv.Atoi: parsing \"cat\": invalid syntax", err.Error())
	})
}
