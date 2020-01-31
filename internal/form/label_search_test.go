package form

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLabelSearchForm(t *testing.T) {
	form := &LabelSearch{}

	assert.IsType(t, new(LabelSearch), form)
}

func TestParseQueryStringLabel(t *testing.T) {

	t.Run("valid query", func(t *testing.T) {
		form := &LabelSearch{Query: "name:cat favorites:true count:10 all:false query:\"query text\""}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Nil(t, err)
		assert.Equal(t, "cat", form.Name)
		assert.Equal(t, true, form.Favorites)
		assert.Equal(t, 10, form.Count)
		assert.Equal(t, false, form.All)
		assert.Equal(t, "query text", form.Query)
	})
	t.Run("valid query 2", func(t *testing.T) {
		form := &LabelSearch{Query: "slug:cat favorites:false offset:2 order:oldest"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}
		assert.Equal(t, "cat", form.Slug)
		assert.Equal(t, false, form.Favorites)
		assert.Equal(t, 2, form.Offset)
		assert.Equal(t, "oldest", form.Order)
	})
	t.Run("valid query with umlauts", func(t *testing.T) {
		form := &LabelSearch{Query: "query:\"tübingen\""}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "tübingen", form.Query)
	})
	t.Run("query for invalid filter", func(t *testing.T) {
		form := &LabelSearch{Query: "xxx:false"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal("err should NOT be nil")
		}

		log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: Xxx", err.Error())
	})
	t.Run("query for favorites with uncommon bool value", func(t *testing.T) {
		form := &LabelSearch{Query: "favorites:0.99"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.True(t, form.Favorites)
	})
	t.Run("query for count with invalid type", func(t *testing.T) {
		form := &LabelSearch{Query: "count:2019-01-15"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal("err should NOT be nil")
		}

		log.Debugf("%+v\n", form)

		assert.Equal(t, "strconv.Atoi: parsing \"2019-01-15\": invalid syntax", err.Error())
	})
}
