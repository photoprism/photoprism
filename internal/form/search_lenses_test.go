package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLensSearchForm(t *testing.T) {
	form := &SearchLenses{}

	assert.IsType(t, new(SearchLenses), form)
}

func TestParseQueryStringLens(t *testing.T) {
	t.Run("ValidQuery", func(t *testing.T) {
		form := &SearchLenses{Query: "name:cat nomake:true q:\"query text\""}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "cat", form.Name)
		assert.Equal(t, true, form.NoMake)
		assert.Equal(t, 0, form.Count)
		assert.Equal(t, "query text", form.Query)
	})
	t.Run("ValidQueryTwo", func(t *testing.T) {
		form := &SearchLenses{Query: "slug:cat nomake:false"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "cat", form.Slug)
		assert.Equal(t, false, form.NoMake)
		assert.Equal(t, 0, form.Count)
		assert.Equal(t, 0, form.Offset)
	})
	t.Run("ValidQueryWithUmlauts", func(t *testing.T) {
		form := &SearchLenses{Query: "q:\"tübingen\""}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "tübingen", form.Query)
	})
	t.Run("QueryForInvalidFilter", func(t *testing.T) {
		form := &SearchLenses{Query: "xxx:false"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal("err should NOT be nil")
		}

		// log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: xxx", err.Error())
	})
	t.Run("QueryForNoMakeWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchLenses{Query: "nomake:0"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.False(t, form.NoMake)
	})
	t.Run("QueryForNoMakeWithInvalidType", func(t *testing.T) {
		form := &SearchLenses{Query: "nomake:2019-01-15"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.NoMake)
	})
}

func TestNewLensSearch(t *testing.T) {
	r := NewLensSearch("cat")
	assert.IsType(t, SearchLenses{}, r)
}
