package forms

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLabelSearchForm(t *testing.T) {
	form := &LabelSearchForm{}

	assert.IsType(t, new(LabelSearchForm), form)
}

func TestParseQueryStringLabel(t *testing.T) {

	t.Run("valid query", func(t *testing.T) {
		form := &LabelSearchForm{Query: "name:cat favorites:true count:10 priority:4 query:\"query text\""}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Nil(t, err)
		assert.Equal(t, "cat", form.Name)
		assert.Equal(t, true, form.Favorites)
		assert.Equal(t, 10, form.Count)
		assert.Equal(t, 4, form.Priority)
		assert.Equal(t, "query text", form.Query)
	})
	t.Run("valid query 2", func(t *testing.T) {
		form := &LabelSearchForm{Query: "slug:cat favorites:false offset:2 order:oldest"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Nil(t, err)
		assert.Equal(t, "cat", form.Slug)
		assert.Equal(t, false, form.Favorites)
		assert.Equal(t, 2, form.Offset)
		assert.Equal(t, "oldest", form.Order)
	})
	t.Run("query for invalid filter", func(t *testing.T) {
		form := &LabelSearchForm{Query: "xxx:false"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: Xxx", err.Error())
	})
	t.Run("query for favorites with invalid type", func(t *testing.T) {
		form := &LabelSearchForm{Query: "favorites:cat"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Equal(t, "not a bool value: Favorites", err.Error())
	})
	t.Run("query for count with invalid type", func(t *testing.T) {
		form := &LabelSearchForm{Query: "count:cat"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Equal(t, "strconv.Atoi: parsing \"cat\": invalid syntax", err.Error())
	})
}
