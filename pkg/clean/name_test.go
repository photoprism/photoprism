package clean

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
		assert.Equal(t, "william HenRy gates'", Name("william \"HenRy\" gates' "))
	})
	t.Run("Slash", func(t *testing.T) {
		assert.Equal(t, "william McCorn / gates'", Name("william\\ \"McCorn\" / gates' "))
	})
	t.Run("SpecialCharacters", func(t *testing.T) {
		assert.Equal(t,
			"'', '', '~', '', '/', '', '', '&', '|', '+', '=', '', Foo '@', '!', '?', ':', '', '', '', McBar '', ''",
			Name("'\"', '`', '~', '\\\\', '/', '*', '%', '&', '|', '+', '=', '$', Foo '@', '!', '?', ':', ';', '<', '>', McBar '{', '}'"),
		)
	})
	t.Run("Chinese", func(t *testing.T) {
		assert.Equal(t, "陈 赵", Name(" 陈  赵"))
	})
	t.Run("Control Character", func(t *testing.T) {
		assert.Equal(t, "William Henry Gates III", Name("William Henry Gates III"+string(rune(1))))
	})
	t.Run("Space", func(t *testing.T) {
		assert.Equal(t, "", Name("        "))
	})
}

func TestNameCapitalized(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", NameCapitalized(""))
	})
	t.Run("BillGates", func(t *testing.T) {
		assert.Equal(t, "William Henry Gates III", NameCapitalized("William Henry Gates III"))
	})
	t.Run("Quotes", func(t *testing.T) {
		assert.Equal(t, "William HenRy Gates'", NameCapitalized("william \"HenRy\" gates' "))
	})
	t.Run("Slash", func(t *testing.T) {
		assert.Equal(t, "William McCorn / Gates'", NameCapitalized("william\\ \"McCorn\" / gates' "))
	})
	t.Run("SpecialCharacters", func(t *testing.T) {
		assert.Equal(t,
			"'', '', '~', '', '/', '', '', '&', '|', '+', '=', '', Foo '@', '!', '?', ':', '', '', '', McBar '', ''",
			Name("'\"', '`', '~', '\\\\', '/', '*', '%', '&', '|', '+', '=', '$', Foo '@', '!', '?', ':', ';', '<', '>', McBar '{', '}'"),
		)
	})
	t.Run("Chinese", func(t *testing.T) {
		assert.Equal(t, "陈 赵", NameCapitalized(" 陈  赵"))
	})
}

func TestDlName(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", DlName(""))
	})
	t.Run("BillGates", func(t *testing.T) {
		assert.Equal(t, "William Henry Gates III", DlName("William Henry Gates III"))
	})
	t.Run("Quotes", func(t *testing.T) {
		assert.Equal(t, "william HenRy gates'", DlName("william \"HenRy\" gates' "))
	})
	t.Run("Ellipsis", func(t *testing.T) {
		assert.Equal(t, "Test - Say goodbye to your adobe subscription 📹 We are better…", DlName("Test - Say goodbye to your adobe subscription 📹 We are better…"))
	})
	t.Run("SpecialCharacters", func(t *testing.T) {
		assert.Equal(t,
			"'', '', '~', '', '', '', '', '&', '', '+', '=', '', Foo '@', '!', '', '', '', '', '', McBar '', ''",
			DlName("'\"', '`', '~', '\\\\', '/', '*', '%', '&', '|', '+', '=', '$', Foo '@', '!', '?', ':', ';', '<', '>', McBar '{', '}'"),
		)
	})
	t.Run("Chinese", func(t *testing.T) {
		assert.Equal(t, "陈 赵", DlName(" 陈  赵"))
	})
	t.Run("Control Character", func(t *testing.T) {
		assert.Equal(t, "William Henry Gates III", DlName("William Henry Gates III"+string(rune(1))))
	})
	t.Run("Space", func(t *testing.T) {
		assert.Equal(t, "", DlName("        "))
	})
}
