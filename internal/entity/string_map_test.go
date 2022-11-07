package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStringMap(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		m := NewStringMap(nil)

		assert.Equal(t, "", m.Get("foo"))
	})
	t.Run("Strings", func(t *testing.T) {
		m := NewStringMap(Strings{"foo": "bar", "bar": "baz"})

		assert.Equal(t, "bar", m.Get("foo"))
		assert.Equal(t, "", m.Get("FOO"))
		assert.Equal(t, "baz", m.Get("bar"))
		assert.Equal(t, "", m.Get("bAr"))
		assert.Equal(t, "", m.Get("baz"))
		assert.Equal(t, "", m.Get(""))
	})
}

func TestStringMap_Set(t *testing.T) {
	t.Run("StartingEmpty", func(t *testing.T) {
		m := NewStringMap(nil)

		assert.Equal(t, "", m.Get("foo"))

		m.Set("foo", "bar")

		assert.Equal(t, "bar", m.Get("foo"))

		m.Set("foo", "bar")

		assert.Equal(t, "bar", m.Get("foo"))

		m.Set("foo", "xxx")

		assert.Equal(t, "xxx", m.Get("foo"))

		m.Set("foo", "")

		assert.Equal(t, "", m.Get("foo"))
	})
	t.Run("WithStrings", func(t *testing.T) {
		m := NewStringMap(Strings{"foo": "bar", "bar": "baz"})

		assert.Equal(t, "bar", m.Get("foo"))

		m.Set("foo", "bar")

		assert.Equal(t, "baz", m.Get("bar"))

		m.Set("bar", "")

		assert.Equal(t, "", m.Get("bar"))

		m.Set("foo", "xxx")

		assert.Equal(t, "xxx", m.Get("foo"))

		m.Set("foo", "")

		assert.Equal(t, "", m.Get("foo"))
	})
}

func TestStringMap_Key(t *testing.T) {
	t.Run("StartingEmpty", func(t *testing.T) {
		m := NewStringMap(nil)

		assert.Equal(t, "", m.Get("foo"))

		m.Set("foo", "bar")
		m.Set("cat", "Windows")
		m.Set("dog", "WINDOWS")
		m.Set("Dog", "WINDOWS")

		assert.Equal(t, "Dog", m.Key("windows"))
		assert.Equal(t, "Dog", m.Key("Windows"))
		assert.Equal(t, "Dog", m.Key("WINDOWS"))
		assert.Equal(t, "bar", m.Get("foo"))

		m.Unset("Dog")

		assert.Equal(t, "dog", m.Key("WINDOWS"))
		assert.Equal(t, "foo", m.Key("bar"))
		assert.Equal(t, "", m.Key("Dog"))
	})
	t.Run("WithStrings", func(t *testing.T) {
		m := NewStringMap(Strings{"foo": "bar", "bar": "baz", "Bar": "Windows"})

		assert.Equal(t, "Bar", m.Key("windows"))
		assert.Equal(t, "Bar", m.Key("Windows"))
		assert.Equal(t, "bar", m.Get("foo"))

		m.Set("Foo", "Bar")
		m.Set("My", "Bar")

		assert.Equal(t, "bar", m.Get("foo"))
		assert.Equal(t, "Bar", m.Get("Foo"))
		assert.Equal(t, "Bar", m.Get("My"))
		assert.Equal(t, "My", m.Key("bar"))

		m.Set("My", "")

		assert.Equal(t, "", m.Get("My"))
		assert.Equal(t, "Foo", m.Key("bar"))
	})
}

func TestStringMap_KeyExists(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, NewStringMap(Strings{"foo": "bar"}).Has("foo"))
		assert.True(t, NewStringMap(Strings{"foo": "bar", "zzz": "bar"}).Has("zzz"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, NewStringMap(Strings{"foo": "bar"}).Has(""))
		assert.False(t, NewStringMap(Strings{"foo": "bar"}).Has("zzz"))
	})
}

func TestStringMap_Missing(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		assert.False(t, NewStringMap(Strings{"foo": "bar"}).Missing("foo"))
		assert.False(t, NewStringMap(Strings{"foo": "bar", "zzz": "bar"}).Missing("zzz"))
	})
	t.Run("True", func(t *testing.T) {
		assert.True(t, NewStringMap(Strings{"foo": "bar"}).Missing(""))
		assert.True(t, NewStringMap(Strings{"foo": "bar"}).Missing("zzz"))
	})
}

func TestStringMap_ValueExists(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, NewStringMap(Strings{"foo": "bar"}).HasValue("bar"))
		assert.True(t, NewStringMap(Strings{"foo": "bar", "zzz": "bar"}).HasValue("bar"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, NewStringMap(Strings{"foo": "bar"}).HasValue(""))
		assert.False(t, NewStringMap(Strings{"foo": "bar"}).HasValue("zzz"))
	})
}

func TestStringMap_Unset(t *testing.T) {
	t.Run("StartingEmpty", func(t *testing.T) {
		m := NewStringMap(nil)

		assert.Equal(t, "", m.Get("foo"))

		m.Unset("foo")

		assert.Equal(t, "", m.Get("foo"))

		m.Set("foo", "xxx")

		assert.Equal(t, "xxx", m.Get("foo"))

		m.Unset("foo")

		assert.Equal(t, "", m.Get("foo"))
		assert.Equal(t, "", m.Get("bar"))
	})
	t.Run("WithStrings", func(t *testing.T) {
		m := NewStringMap(Strings{"foo": "bar", "bar": "baz"})

		assert.Equal(t, "bar", m.Get("foo"))

		m.Unset("foo")

		assert.Equal(t, "", m.Get("foo"))

		m.Set("foo", "xxx")

		assert.Equal(t, "xxx", m.Get("foo"))

		m.Unset("foo")

		assert.Equal(t, "", m.Get("foo"))
		assert.Equal(t, "baz", m.Get("bar"))
	})
}
