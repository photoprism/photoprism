package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFaceSearch(t *testing.T) {
	r := NewFaceSearch("yes")
	assert.IsType(t, SearchFaces{}, r)
}

func TestFaceSearch_GetQuery(t *testing.T) {
	form := &SearchFaces{Query: "test"}

	assert.Equal(t, "test", form.GetQuery())
}

func TestFaceSearch_SetQuery(t *testing.T) {
	form := &SearchFaces{Query: "test"}
	form.SetQuery("new query")

	assert.Equal(t, "new query", form.GetQuery())
}

func TestFaceSearch_ParseQueryString(t *testing.T) {
	t.Run("valid query", func(t *testing.T) {
		form := &SearchFaces{Query: "subject:test"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "test", form.Subject)
		assert.Equal(t, 0, form.Count)
		assert.Equal(t, 0, form.Offset)

	})
	t.Run("valid query with umlauts", func(t *testing.T) {
		form := &SearchFaces{Query: "q:\"tübingen\""}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "tübingen", form.Query)
	})
	t.Run("query for invalid filter", func(t *testing.T) {
		form := &SearchFaces{Query: "xxx:false"}

		err := form.ParseQueryString()

		if err == nil {
			t.FailNow()
		}

		// log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: xxx", err.Error())
	})
	t.Run("query for count with invalid type", func(t *testing.T) {
		form := &SearchFaces{Query: "count:cat"}

		err := form.ParseQueryString()

		assert.Error(t, err)
	})
}
