package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchServices(t *testing.T) {
	form := &SearchServices{}

	assert.IsType(t, new(SearchServices), form)
}

func TestSearchServices_GetQuery(t *testing.T) {
	form := &SearchServices{Query: "q: webdav, share:true, sync:false, status:test"}

	r := form.GetQuery()
	assert.Equal(t, "q: webdav, share:true, sync:false, status:test", r)
}

func TestSearchServices_SetQuery(t *testing.T) {
	form := &SearchServices{}
	assert.Equal(t, "", form.GetQuery())
	form.SetQuery("query test")
	assert.Equal(t, "query test", form.GetQuery())
}

func TestSearchServices_ParseQueryString(t *testing.T) {

	t.Run("valid query", func(t *testing.T) {
		form := &SearchServices{Query: "q: webdäv share:true sync:false status:test"}

		err := form.ParseQueryString()

		if err != nil {
			t.FailNow()
		}

		// log.Debugf("%+v\n", form)

		assert.Equal(t, "webdäv", form.Query)
		assert.Equal(t, true, form.Share)
		assert.Equal(t, false, form.Sync)
		assert.Equal(t, 0, form.Count)
	})

	t.Run("query for invalid filter", func(t *testing.T) {
		form := &SearchServices{Query: "xxx:false"}

		err := form.ParseQueryString()

		if err == nil {
			t.FailNow()
		}

		// log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: xxx", err.Error())
	})
	t.Run("query for sync with uncommon bool value", func(t *testing.T) {
		form := &SearchServices{Query: "sync:cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.True(t, form.Sync)
	})
	t.Run("query for count with invalid type", func(t *testing.T) {
		form := &SearchAlbums{Query: "count:cat"}

		err := form.ParseQueryString()

		assert.Error(t, err)
	})
}

func TestNewSearchServices(t *testing.T) {
	r := NewSearchServices("holiday")
	assert.IsType(t, SearchServices{}, r)
}
