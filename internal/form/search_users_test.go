package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchUsers_GetQuery(t *testing.T) {
	form := &SearchUsers{Query: "John Doe"}

	assert.Equal(t, "John Doe", form.GetQuery())
}

func TestSearchUsers_SetQuery(t *testing.T) {
	form := &SearchUsers{Query: "John Doe"}
	form.SetQuery("Jane")

	assert.Equal(t, "Jane", form.GetQuery())
}

func TestSearchUsers_ParseQueryString(t *testing.T) {
	form := &SearchUsers{Query: "John Doe", Email: "john@test.com", Name: "John"}

	err := form.ParseQueryString()

	if err != nil {
		t.Fatal("err should be nil")
	}

	assert.Equal(t, "john@test.com", form.Email)
	assert.Equal(t, "john doe", form.Query)
	assert.Equal(t, "John", form.Name)
}
