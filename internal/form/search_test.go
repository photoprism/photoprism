package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQueryString_Bindable(t *testing.T) {
	t.Run("FieldsBindable", func(t *testing.T) {
		f := &SearchPhotos{Query: "cat\x00", Hash: "ab\x00cd", Country: "d\x00e", Camera: "Can\xffon", Title: "Grüße 日本"}
		assert.NoError(t, ParseQueryString(f))
		assert.Equal(t, "cat", f.Query)
		assert.Equal(t, "abcd", f.Hash)
		assert.Equal(t, "de", f.Country)
		assert.Equal(t, "Canon", f.Camera)
		assert.Equal(t, "Grüße 日本", f.Title)
	})
	t.Run("QueryValues", func(t *testing.T) {
		f := &SearchPhotos{Query: "title:\"a\x00b\" label:cat"}
		assert.NoError(t, ParseQueryString(f))
		assert.Equal(t, "ab", f.Title)
		assert.Equal(t, "cat", f.Label)
	})
}

func TestBindableStrings(t *testing.T) {
	t.Run("Struct", func(t *testing.T) {
		f := &SearchAlbums{Query: "a\x00b", Category: "\x00", Location: "Berlin"}
		bindableStrings(f)
		assert.Equal(t, "ab", f.Query)
		assert.Equal(t, "", f.Category)
		assert.Equal(t, "Berlin", f.Location)
	})
	t.Run("NotPointer", func(t *testing.T) {
		f := SearchAlbums{Query: "a\x00b"}
		bindableStrings(f)
		bindableStrings(nil)
		bindableStrings((*SearchAlbums)(nil))
		s := "a\x00b"
		bindableStrings(&s)
		assert.Equal(t, "a\x00b", s)
	})
}
