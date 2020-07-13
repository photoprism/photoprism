package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestData_AddKeyword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data := NewData()

		assert.Equal(t, "", data.Keywords)

		data.AddKeyword("FooBar")

		assert.Equal(t, "foobar", data.Keywords)

		data.AddKeyword("BAZ")

		assert.Equal(t, "foobar, baz", data.Keywords)
	})

	t.Run("ignore", func(t *testing.T) {
		data := NewData()

		assert.Equal(t, "", data.Keywords)

		data.AddKeyword("Fo")

		assert.Equal(t, "", data.Keywords)
	})
}

func TestData_AutoAddKeywords(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data := NewData()

		assert.Equal(t, "", data.Keywords)

		data.AutoAddKeywords("FooBar burst baz flash")

		assert.Equal(t, "burst", data.Keywords)
	})

	t.Run("ignore", func(t *testing.T) {
		data := NewData()

		assert.Equal(t, "", data.Keywords)

		data.AutoAddKeywords("FooBar go pro baz banana")

		assert.Equal(t, "", data.Keywords)
	})
}
