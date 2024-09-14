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
		assert.Equal(t, Attr{{Key: "bar", Value: "true"}, {Key: "baz", Value: "true"}, {Key: "foo", Value: "true"}}, attr.Sort())
	})
	t.Run("WhitespaceKeys", func(t *testing.T) {
		attr := ParseAttr(" foo         bar  baz    ")

		assert.Len(t, attr, 3)
		assert.Equal(t, Attr{{Key: "foo", Value: "true"}, {Key: "bar", Value: "true"}, {Key: "baz", Value: "true"}}, attr)
		assert.Equal(t, Attr{{Key: "bar", Value: "true"}, {Key: "baz", Value: "true"}, {Key: "foo", Value: "true"}}, attr.Sort())
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
		assert.Equal(t,
			Attr{
				{Key: "BIG", Value: "true"},
				{Key: "CAT", Value: "FISH"},
				{Key: "bar", Value: "false"},
				{Key: "baz", Value: "true"},
				{Key: "berghain", Value: "berlin:germany"},
				{Key: "biZZ", Value: "false"},
				{Key: "foo", Value: "true"},
				{Key: "hello", Value: "false"},
			}, attr.Sort(),
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
		assert.Equal(t, Attr{
			{Key: "albums", Value: "true"},
			{Key: "config.view", Value: "false"},
			{Key: "files", Value: "true"},
			{Key: "files.read", Value: "true"},
			{Key: "people.view", Value: "true"},
			{Key: "photos", Value: "true"},
			{Key: "photos.create", Value: "false"},
		}, attr.Sort())
	})
}

func TestParseKeyValue(t *testing.T) {
	t.Run("Any", func(t *testing.T) {
		kv := ParseKeyValue("*")

		assert.Equal(t, "*", kv.Key)
		assert.Equal(t, "", kv.Value)
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
		s := "  admin.conversations.removeCustomRetention  * admin.usergroups:read  me:yes FOOt0-2U	6VU #$#%$ cm,Nu"
		attr := ParseAttr(s)

		assert.Len(t, attr, 7)
		assert.Equal(t, "6VU FOOt0-2U admin.conversations.removeCustomRetention admin.usergroups:read cmNu me *", attr.String())
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

func TestAttr_Find(t *testing.T) {
	t.Run("Any", func(t *testing.T) {
		s := "*"
		attr := ParseAttr(s)

		assert.Len(t, attr, 1)
		result := attr.Find("metrics")

		assert.Equal(t, All, result.Key)
		assert.Equal(t, "", result.Value)
	})
	t.Run("Empty", func(t *testing.T) {
		s := "*"
		attr := ParseAttr(s)

		assert.Len(t, attr, 1)
		result := attr.Find("")
		assert.Equal(t, "", result.Key)
		assert.Equal(t, "", result.Value)
	})
	t.Run("All", func(t *testing.T) {
		s := "*"
		attr := ParseAttr(s)

		assert.Len(t, attr, 1)
		result := attr.Find("*")
		assert.Equal(t, All, result.Key)
		assert.Equal(t, "", result.Value)
	})
	t.Run("ValueAll", func(t *testing.T) {
		s := "*"
		attr := ParseAttr(s)

		assert.Len(t, attr, 1)
		result := attr.Find("6VU:*")
		assert.Equal(t, All, result.Key)
		assert.Equal(t, "", result.Value)
	})
	t.Run("Scopes", func(t *testing.T) {
		attr := ParseAttr("files files.read:true photos photos.create:false albums:true people.view:true config.view:false")

		assert.Len(t, attr, 7)
		result := attr.Find("people.view")
		assert.Equal(t, "people.view", result.Key)
		assert.Equal(t, True, result.Value)
	})
	t.Run("ReadAll", func(t *testing.T) {
		s := "read *"
		attr := ParseAttr(s)

		assert.Len(t, attr, 2)
		result := attr.Find("read")
		assert.Equal(t, "read", result.Key)
		assert.Equal(t, True, result.Value)
	})
	t.Run("ReadFalse", func(t *testing.T) {
		s := "read:false *"
		attr := ParseAttr(s)

		assert.Len(t, attr, 2)
		result := attr.Find("read:*")
		assert.Equal(t, "read", result.Key)
		assert.Equal(t, False, result.Value)
		result = attr.Find("read:false")
		assert.Equal(t, "read", result.Key)
		assert.Equal(t, False, result.Value)
	})
	t.Run("ReadOther", func(t *testing.T) {
		s := "read:other *"
		attr := ParseAttr(s)

		assert.Len(t, attr, 2)

		result := attr.Find("read")
		assert.Equal(t, All, result.Key)
		assert.Equal(t, "", result.Value)

		result = attr.Find("read:other")
		assert.Equal(t, "read", result.Key)
		assert.Equal(t, "other", result.Value)

		result = attr.Find("read:true")
		assert.Equal(t, All, result.Key)
		assert.Equal(t, "", result.Value)

		result = attr.Find("read:false")
		assert.Equal(t, "", result.Key)
		assert.Equal(t, "", result.Value)
	})
}
