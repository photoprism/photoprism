package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelSearchForm(t *testing.T) {
	form := &SearchLabels{}

	assert.IsType(t, new(SearchLabels), form)
}

func TestParseQueryStringLabel(t *testing.T) {
	t.Run("valid query", func(t *testing.T) {
		form := &SearchLabels{Query: "name:cat favorite:true all:false q:\"query text\""}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "cat", form.Name)
		assert.Equal(t, true, form.Favorite)
		assert.Equal(t, 0, form.Count)
		assert.Equal(t, false, form.All)
		assert.Equal(t, "query text", form.Query)
	})
	t.Run("valid query 2", func(t *testing.T) {
		form := &SearchLabels{Query: "slug:cat favorite:false"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "cat", form.Slug)
		assert.Equal(t, false, form.Favorite)
		assert.Equal(t, 0, form.Count)
		assert.Equal(t, 0, form.Offset)
		assert.Equal(t, "", form.Order)
	})
	t.Run("valid query with umlauts", func(t *testing.T) {
		form := &SearchLabels{Query: "q:\"tübingen\""}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "tübingen", form.Query)
	})
	t.Run("query for invalid filter", func(t *testing.T) {
		form := &SearchLabels{Query: "xxx:false"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal("err should NOT be nil")
		}

		// log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: xxx", err.Error())
	})
	t.Run("query for favorites with uncommon bool value", func(t *testing.T) {
		form := &SearchLabels{Query: "favorite:0.99"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.False(t, form.Favorite)
	})
	t.Run("query for count with invalid type", func(t *testing.T) {
		form := &SearchLabels{Query: "all:2019-01-15"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.All)
	})
}

func TestNewLabelSearch(t *testing.T) {
	r := NewLabelSearch("cat")
	assert.IsType(t, SearchLabels{}, r)
}
