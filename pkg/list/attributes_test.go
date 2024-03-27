package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAttr(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		attr := ParseAttr("")

		assert.Len(t, attr, 0)
		assert.Equal(t, Attr{}, attr)
	})
	t.Run("Keys", func(t *testing.T) {
		attr := ParseAttr("foo bar baz")

		assert.Len(t, attr, 3)
		assert.Equal(t, Attr{{Key: "foo", Value: "true"}, {Key: "bar", Value: "true"}, {Key: "baz", Value: "true"}}, attr)
	})
	t.Run("WhitespaceKeys", func(t *testing.T) {
		attr := ParseAttr(" foo         bar  baz    ")

		assert.Len(t, attr, 3)
		assert.Equal(t, Attr{{Key: "foo", Value: "true"}, {Key: "bar", Value: "true"}, {Key: "baz", Value: "true"}}, attr)
	})
	t.Run("Values", func(t *testing.T) {
		attr := ParseAttr("foo:yes bar:disable baz:true    biZZ:false   BIG CAT:FISH berghain:berlin:germany hello:off")

		assert.Len(t, attr, 8)
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
			}, attr,
		)
	})
	t.Run("Scopes", func(t *testing.T) {
		attr := ParseAttr("files files.read:true photos photos.create:false albums:true people.view:true config.view:false")

		assert.Len(t, attr, 7)
		assert.Equal(t, Attr{
			{Key: "files", Value: "true"},
			{Key: "files.read", Value: "true"},
			{Key: "photos", Value: "true"},
			{Key: "photos.create", Value: "false"},
			{Key: "albums", Value: "true"},
			{Key: "people.view", Value: "true"},
			{Key: "config.view", Value: "false"},
		}, attr)
	})
}

func TestParseKeyValue(t *testing.T) {
	t.Run("Any", func(t *testing.T) {
		kv := ParseKeyValue("*")

		assert.Equal(t, "*", kv.Key)
		assert.Equal(t, "true", kv.Value)
	})
	t.Run("Scope", func(t *testing.T) {
		kv := ParseKeyValue("files.read:true")

		t.Logf("Scope: %#v", kv)
		assert.Equal(t, &KeyValue{Key: "files.read", Value: "true"}, kv)
	})
}

func TestAttr_String(t *testing.T) {
	t.Run("SlackScope", func(t *testing.T) {
		s := "admin.conversations.removeCustomRetention admin.usergroups:read"
		attr := ParseAttr(s)

		expected := Attr{
			{Key: "admin.conversations.removeCustomRetention", Value: "true"},
			{Key: "admin.usergroups", Value: "read"},
		}

		assert.Len(t, attr, 2)
		assert.Equal(t, expected, attr)
		assert.Equal(t, s, attr.String())
	})
	t.Run("Random", func(t *testing.T) {
		s := "  admin.conversations.removeCustomRetention   admin.usergroups:read  me:yes FOOt0-2U	6VU #$#%$ cm,Nu"
		attr := ParseAttr(s)

		assert.Len(t, attr, 6)
		assert.Equal(t, "6VU FOOt0-2U admin.conversations.removeCustomRetention admin.usergroups:read cmNu me", attr.String())
	})
}

func TestAttr_Contains(t *testing.T) {
	t.Run("Any", func(t *testing.T) {
		s := "*"
		attr := ParseAttr(s)

		assert.Len(t, attr, 1)
		assert.True(t, attr.Contains("metrics"))
	})
	t.Run("Empty", func(t *testing.T) {
		s := "*"
		attr := ParseAttr(s)

		assert.Len(t, attr, 1)
		assert.False(t, attr.Contains(""))
	})
	t.Run("All", func(t *testing.T) {
		s := "*"
		attr := ParseAttr(s)

		assert.Len(t, attr, 1)
		assert.True(t, attr.Contains("*"))
	})
	t.Run("ValueAll", func(t *testing.T) {
		s := "*"
		attr := ParseAttr(s)

		assert.Len(t, attr, 1)
		assert.True(t, attr.Contains("6VU:*"))
	})
	t.Run("Scopes", func(t *testing.T) {
		attr := ParseAttr("files files.read:true photos photos.create:false albums:true people.view:true config.view:false")

		assert.Len(t, attr, 7)
		assert.True(t, attr.Contains("people.view"))
	})
}
