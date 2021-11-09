package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpaced(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "  ", Spaced(""))
	})
	t.Run("Space", func(t *testing.T) {
		assert.Equal(t, "   ", Spaced(" "))
	})
	t.Run("Chinese", func(t *testing.T) {
		assert.Equal(t, " 李 ", Spaced("李"))
	})
	t.Run("And", func(t *testing.T) {
		assert.Equal(t, " and ", Spaced("and"))
	})
}

func TestStripOr(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", StripOr(""))
	})
	t.Run("EnOr", func(t *testing.T) {
		assert.Equal(t, "or", StripOr("or"))
	})
	t.Run("SpacedEnOr", func(t *testing.T) {
		assert.Equal(t, "李 or Foo", StripOr("李 or Foo"))
	})
	t.Run("Or", func(t *testing.T) {
		assert.Equal(t, "李   Foo", StripOr("李 | Foo"))
	})
}

func TestQueryTooShort(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, QueryTooShort(""))
	})
	t.Run("IsTooShort", func(t *testing.T) {
		assert.True(t, QueryTooShort("aa"))
	})
	t.Run("Chinese", func(t *testing.T) {
		assert.False(t, QueryTooShort("李"))
	})
	t.Run("OK", func(t *testing.T) {
		assert.False(t, QueryTooShort("foo"))
	})
}
