package form

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestAccountSearchForm(t *testing.T) {
	form := &AccountSearch{}

	assert.IsType(t, new(AccountSearch), form)
}

func TestAccountSearch_GetQuery(t *testing.T) {
	form := &AccountSearch{Query: "query: webdav, share:true, sync:false, status:test, count:10"}

	r := form.GetQuery()
	assert.Equal(t, "query: webdav, share:true, sync:false, status:test, count:10", r)
}

func TestAccountSearch_SetQuery(t *testing.T) {
	form := &AccountSearch{}
	assert.Equal(t, "", form.GetQuery())
	form.SetQuery("query test")
	assert.Equal(t, "query test", form.GetQuery())
}

func TestAccountSearch_ParseQueryString(t *testing.T) {

	t.Run("valid query", func(t *testing.T) {
		form := &AccountSearch{Query: "query: webdäv share:true sync:false status:test count:10"}

		err := form.ParseQueryString()

		if err != nil {
			t.FailNow()
		}

		log.Debugf("%+v\n", form)

		assert.Equal(t, "webdäv", form.Query)
		assert.Equal(t, true, form.Share)
		assert.Equal(t, false, form.Sync)
		assert.Equal(t, 10, form.Count)
	})

	t.Run("query for invalid filter", func(t *testing.T) {
		form := &AccountSearch{Query: "xxx:false"}

		err := form.ParseQueryString()

		if err == nil {
			t.FailNow()
		}

		log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: Xxx", err.Error())
	})
	t.Run("query for sync with uncommon bool value", func(t *testing.T) {
		form := &AccountSearch{Query: "sync:cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.True(t, form.Sync)
	})
	t.Run("query for count with invalid type", func(t *testing.T) {
		form := &AlbumSearch{Query: "count:cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.FailNow()
		}

		log.Debugf("%+v\n", form)

		assert.Equal(t, "strconv.Atoi: parsing \"cat\": invalid syntax", err.Error())
	})
}

func TestNewAccountSearch(t *testing.T) {
	r := NewAccountSearch("holiday")
	assert.IsType(t, AccountSearch{}, r)
}
