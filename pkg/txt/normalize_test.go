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
		result := NormalizeState("Berlin", "de")
		assert.Equal(t, "Berlin", result)
	})

	t.Run("WA", func(t *testing.T) {
		result := NormalizeState("WA", "us")
		assert.Equal(t, "Washington", result)
	})

	t.Run("QCUnknownCountry", func(t *testing.T) {
		result := NormalizeState("QC", "")
		assert.Equal(t, "QC", result)
	})

	t.Run("QCCanada", func(t *testing.T) {
		result := NormalizeState("QC", "ca")
		assert.Equal(t, "Quebec", result)
	})

	t.Run("QCUnitedStates", func(t *testing.T) {
		result := NormalizeState("QC", "us")
		assert.Equal(t, "QC", result)
	})

	t.Run("Wa", func(t *testing.T) {
		result := NormalizeState("Wa", "us")
		assert.Equal(t, "Wa", result)
	})

	t.Run("Washington", func(t *testing.T) {
		result := NormalizeState("Washington", "us")
		assert.Equal(t, "Washington", result)
	})

	t.Run("Never mind Nirvana", func(t *testing.T) {
		result := NormalizeState("Never mind Nirvana.", "us")
		assert.Equal(t, "Never mind Nirvana.", result)
	})

	t.Run("Empty", func(t *testing.T) {
		result := NormalizeState("", "us")
		assert.Equal(t, "", result)
	})

	t.Run("Unknown", func(t *testing.T) {
		result := NormalizeState("zz", "us")
		assert.Equal(t, "", result)
	})

	t.Run("Space", func(t *testing.T) {
		result := NormalizeState(" ", "us")
		assert.Equal(t, "", result)
	})

}
func TestNormalizeQuery(t *testing.T) {
	t.Run("Replace", func(t *testing.T) {
		q := NormalizeQuery("table spoon & usa | img% json OR BILL!")
		assert.Equal(t, "table spoon & usa | img* json|bill", q)
	})
}

func TestNormalizeUsername(t *testing.T) {
	t.Run("Admin ", func(t *testing.T) {
		assert.Equal(t, "admin", NormalizeUsername("Admin "))
	})
	t.Run(" Admin ", func(t *testing.T) {
		assert.Equal(t, "admin", NormalizeUsername(" Admin "))
	})
	t.Run(" admin ", func(t *testing.T) {
		assert.Equal(t, "admin", NormalizeUsername(" admin "))
	})
}
