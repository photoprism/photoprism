package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchSessions_GetQuery(t *testing.T) {
	form := &SearchSessions{Query: "test"}

	assert.Equal(t, "test", form.GetQuery())
}

func TestSearchSessions_SetQuery(t *testing.T) {
	form := &SearchSessions{Query: "test"}
	form.SetQuery("new query")

	assert.Equal(t, "new query", form.GetQuery())
}

func TestSearchSessions_ParseQueryString(t *testing.T) {
	form := &SearchSessions{Query: "test", Count: 3}

	err := form.ParseQueryString()

	if err != nil {
		t.Fatal("err should be nil")
	}

	assert.Equal(t, 3, form.Count)
	assert.Equal(t, "test", form.Query)
}
