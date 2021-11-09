package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeName(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", NormalizeName(""))
	})
	t.Run("BillGates", func(t *testing.T) {
		assert.Equal(t, "William Henry Gates III", NormalizeName("William Henry Gates III"))
	})
	t.Run("Quotes", func(t *testing.T) {
		assert.Equal(t, "William HenRy Gates'", NormalizeName("william \"HenRy\" gates' "))
	})
	t.Run("Slash", func(t *testing.T) {
		assert.Equal(t, "William McCorn Gates'", NormalizeName("william\\ \"McCorn\" / gates' "))
	})
	t.Run("SpecialCharacters", func(t *testing.T) {
		assert.Equal(t,
			"'', '', '', '', '', '', '', '', '', '', '', '', Foo '', '', '', '', '', '', '', McBar '', ''",
			NormalizeName("'\"', '`', '~', '\\\\', '/', '*', '%', '&', '|', '+', '=', '$', Foo '@', '!', '?', ':', ';', '<', '>', McBar '{', '}'"),
		)
	})
	t.Run("Chinese", func(t *testing.T) {
		assert.Equal(t, "陈 赵", NormalizeName(" 陈  赵"))
	})
}

func TestNormalizeState(t *testing.T) {
	t.Run("Berlin", func(t *testing.T) {
		result := NormalizeState("Berlin")
		assert.Equal(t, "Berlin", result)
	})

	t.Run("WA", func(t *testing.T) {
		result := NormalizeState("WA")
		assert.Equal(t, "Washington", result)
	})

	t.Run("Wa", func(t *testing.T) {
		result := NormalizeState("Wa")
		assert.Equal(t, "Wa", result)
	})

	t.Run("Washington", func(t *testing.T) {
		result := NormalizeState("Washington")
		assert.Equal(t, "Washington", result)
	})

	t.Run("Never mind Nirvana", func(t *testing.T) {
		result := NormalizeState("Never mind Nirvana.")
		assert.Equal(t, "Never Mind Nirvana.", result)
	})

	t.Run("Empty", func(t *testing.T) {
		result := NormalizeState("")
		assert.Equal(t, "", result)
	})

	t.Run("Unknown", func(t *testing.T) {
		result := NormalizeState("zz")
		assert.Equal(t, "", result)
	})

	t.Run("Space", func(t *testing.T) {
		result := NormalizeState(" ")
		assert.Equal(t, "", result)
	})

}
func TestNormalizeQuery(t *testing.T) {
	t.Run("Replace", func(t *testing.T) {
		q := NormalizeQuery("table spoon & usa | img% json OR BILL!")
		assert.Equal(t, "table spoon & usa | img* json|bill", q)
	})
}
