package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAttr(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		f := ParseAttr("")
		assert.Len(t, f, 0)
		assert.Equal(t, Attr{}, f)
	})
	t.Run("Keys", func(t *testing.T) {
		f := ParseAttr("foo bar baz")
		assert.Len(t, f, 3)
		assert.Equal(t, Attr{{Key: "foo", Value: "true"}, {Key: "bar", Value: "true"}, {Key: "baz", Value: "true"}}, f)
	})
	t.Run("WhitespaceKeys", func(t *testing.T) {
		f := ParseAttr(" foo         bar  baz    ")
		assert.Len(t, f, 3)
		assert.Equal(t, Attr{{Key: "foo", Value: "true"}, {Key: "bar", Value: "true"}, {Key: "baz", Value: "true"}}, f)
	})
	t.Run("Values", func(t *testing.T) {
		f := ParseAttr("foo:yes bar:disable baz:true    biZZ:false   BIG CAT:FISH berghain:berlin:germany hello:off")
		assert.Len(t, f, 8)
		assert.Equal(t,
			Attr{
				{Key: "foo", Value: "true"},
				{Key: "bar", Value: "false"},
				{Key: "baz", Value: "true"},
				{Key: "biZZ", Value: "false"},
				{Key: "BIG", Value: "true"},
				{Key: "CAT", Value: "FISH"},
				{Key: "berghain", Value: "berlin:germany"},
				{Key: "hello", Value: "false"},
			}, f,
		)
	})
}

func TestAttr_String(t *testing.T) {
	t.Run("SlackScope", func(t *testing.T) {
		s := "admin.conversations.removeCustomRetention admin.usergroups:read"
		f := ParseAttr(s)

		expected := Attr{
			{Key: "admin.conversations.removeCustomRetention", Value: "true"},
			{Key: "admin.usergroups", Value: "read"},
		}

		assert.Len(t, f, 2)
		assert.Equal(t, expected, f)
		assert.Equal(t, s, f.String())
	})
	t.Run("Random", func(t *testing.T) {
		s := "  admin.conversations.removeCustomRetention   admin.usergroups:read  me:yes FOOt0-2U	6VU #$#%$ cm,Nu"
		f := ParseAttr(s)

		assert.Len(t, f, 6)
		assert.Equal(t, "6VU FOOt0-2U admin.conversations.removeCustomRetention admin.usergroups:read cmNu me", f.String())
	})
}

func TestAttr_Contains(t *testing.T) {
	t.Run("Any", func(t *testing.T) {
		s := "*"
		a := ParseAttr(s)

		assert.Len(t, a, 1)

		t.Logf("Contains: %s", a[0])
		assert.True(t, a.Contains("metrics"))
	})
}

func TestParseKeyValue(t *testing.T) {
	t.Run("Any", func(t *testing.T) {
		v := ParseKeyValue("*")
		t.Logf("Key: '%s'", v.Key)
		t.Logf("Value: '%s'", v.Value)
		assert.Equal(t, "*", v.Key)
		assert.Equal(t, "true", v.Value)
	})
}
