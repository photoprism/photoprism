package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeQuery(t *testing.T) {
	t.Run("Replace", func(t *testing.T) {
		q := NormalizeQuery("table spoon & usa | img% json OR BILL!")
		assert.Equal(t, "table spoon & usa | img* json|bill", q)
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
