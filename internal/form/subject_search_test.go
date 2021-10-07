package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubjectSearchForm(t *testing.T) {
	form := &SubjectSearch{}

	assert.IsType(t, new(SubjectSearch), form)
}

func TestParseQueryStringSubject(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		form := &SubjectSearch{Query: "type:person favorite:true hidden:all count:5"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, "person", form.Type)
		assert.Equal(t, "true", form.Favorite)
		assert.Equal(t, "all", form.Hidden)
		assert.Equal(t, 5, form.Count)
	})
}

func TestNewSubjectSearch(t *testing.T) {
	r := NewSubjectSearch("john")
	assert.IsType(t, SubjectSearch{}, r)
}
