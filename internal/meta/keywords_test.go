package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestData_AddKeywords(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data := NewData()

		assert.Equal(t, "", data.Keywords.String())

		data.AddKeywords("FooBar")

		assert.Equal(t, "foobar", data.Keywords.String())

		data.AddKeywords("BAZ; pro")

		assert.Equal(t, "baz, foobar, pro", data.Keywords.String())
	})

	t.Run("ignore", func(t *testing.T) {
		data := NewData()

		assert.Equal(t, "", data.Keywords.String())

		data.AddKeywords("Fo")

		assert.Equal(t, "fo", data.Keywords.String())
	})
}

func TestData_AutoAddKeywords(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data := NewData()

		assert.Equal(t, "", data.Keywords.String())

		data.AutoAddKeywords("FooBar burst baz flash")

		assert.Equal(t, "burst", data.Keywords.String())
	})

	t.Run("ignore", func(t *testing.T) {
		data := NewData()

		assert.Equal(t, "", data.Keywords.String())

		data.AutoAddKeywords("FooBar go pro baz banana")

		assert.Equal(t, "", data.Keywords.String())
	})

	t.Run("ignore because too short", func(t *testing.T) {
		data := NewData()

		assert.Equal(t, "", data.Keywords.String())

		data.AutoAddKeywords("es")

		assert.Equal(t, "", data.Keywords.String())
	})
}
