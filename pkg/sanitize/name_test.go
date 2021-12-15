package sanitize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Name(""))
	})
	t.Run("BillGates", func(t *testing.T) {
		assert.Equal(t, "William Henry Gates III", Name("William Henry Gates III"))
	})
	t.Run("Quotes", func(t *testing.T) {
		assert.Equal(t, "William HenRy Gates'", Name("william \"HenRy\" gates' "))
	})
	t.Run("Slash", func(t *testing.T) {
		assert.Equal(t, "William McCorn Gates'", Name("william\\ \"McCorn\" / gates' "))
	})
	t.Run("SpecialCharacters", func(t *testing.T) {
		assert.Equal(t,
			"'', '', '', '', '', '', '', '', '', '', '', '', Foo '', '', '', '', '', '', '', McBar '', ''",
			Name("'\"', '`', '~', '\\\\', '/', '*', '%', '&', '|', '+', '=', '$', Foo '@', '!', '?', ':', ';', '<', '>', McBar '{', '}'"),
		)
	})
	t.Run("Chinese", func(t *testing.T) {
		assert.Equal(t, "陈 赵", Name(" 陈  赵"))
	})
}
