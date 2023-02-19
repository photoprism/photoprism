package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFolderSearchForm(t *testing.T) {
	form := &SearchFolders{}

	assert.IsType(t, new(SearchFolders), form)
}

func TestFolderSearch_GetQuery(t *testing.T) {
	form := &SearchFolders{Query: "test"}

	assert.Equal(t, "test", form.GetQuery())
}

func TestFolderSearch_SetQuery(t *testing.T) {
	form := &SearchFolders{Query: "test"}
	form.SetQuery("new query")

	assert.Equal(t, "new query", form.GetQuery())
}

func TestFolderSearch_Serialize(t *testing.T) {
	form := &SearchFolders{Query: "test", Files: true}

	assert.Equal(t, "q:test files:true", form.Serialize())
}

func TestFolderSearch_SerializeAll(t *testing.T) {
	form := &SearchFolders{Query: "test", Files: true}

	assert.Equal(t, "q:test files:true", form.SerializeAll())
}

func TestParseQueryStringFolder(t *testing.T) {
	t.Run("valid query", func(t *testing.T) {
		form := &SearchFolders{Query: "uncached:false files:true recursive:true"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, false, form.Uncached)
		assert.Equal(t, true, form.Recursive)
		assert.Equal(t, true, form.Files)
		assert.Equal(t, 0, form.Count)
		assert.Equal(t, 0, form.Offset)

	})
	t.Run("valid query with umlauts", func(t *testing.T) {
		form := &SearchFolders{Query: "q:\"tübingen\""}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "tübingen", form.Query)
	})
	t.Run("query for invalid filter", func(t *testing.T) {
		form := &SearchFolders{Query: "xxx:false"}

		err := form.ParseQueryString()

		if err == nil {
			t.FailNow()
		}

		// log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: xxx", err.Error())
	})
	t.Run("query for recursive with uncommon bool value", func(t *testing.T) {
		form := &SearchFolders{Query: "recursive:cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.True(t, form.Recursive)
	})
	t.Run("query for count with invalid type", func(t *testing.T) {
		form := &SearchFolders{Query: "count:cat"}

		err := form.ParseQueryString()

		assert.Error(t, err)
	})
}
