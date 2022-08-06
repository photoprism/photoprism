package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubjectSearchForm(t *testing.T) {
	form := &SearchSubjects{}

	assert.IsType(t, new(SearchSubjects), form)
}

func TestParseQueryStringSubject(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		form := &SearchSubjects{Query: "type:person favorite:true hidden:all"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "person", form.Type)
		assert.Equal(t, "true", form.Favorite)
		assert.Equal(t, "all", form.Hidden)
		assert.Equal(t, 0, form.Count)
	})
}

func TestNewSubjectSearch(t *testing.T) {
	r := NewSubjectSearch("john")
	assert.IsType(t, SearchSubjects{}, r)
}
