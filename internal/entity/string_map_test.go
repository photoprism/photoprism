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

		assert.Equal(t, "", m.Key("WINDOWS"))
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
		assert.Equal(t, "", m.Key("bar"))
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
