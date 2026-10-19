package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	t.Run("OpenQuote", func(t *testing.T) {
		// A quote that is not closed extends the value to the end of the query.
		f := &SearchPhotos{Query: "label:\"cat"}
		require.NoError(t, ParseQueryString(f))
		assert.Equal(t, "cat", f.Label)
		f = &SearchPhotos{Query: "title:x label:\"cat path:dog"}
		require.NoError(t, ParseQueryString(f))
		assert.Equal(t, "x", f.Title)
		assert.Equal(t, "cat path:dog", f.Label)
		assert.Equal(t, "", f.Path)
		f = &SearchPhotos{Query: "\"cat dog"}
		require.NoError(t, ParseQueryString(f))
		assert.Equal(t, "cat dog", f.Query)
		f = &SearchPhotos{Query: "quality:\"3"}
		require.NoError(t, ParseQueryString(f))
		assert.Equal(t, 3, f.Quality)
		assert.ErrorIs(t, ParseQueryString(&SearchPhotos{Query: "path:\"\xff"}), ErrNoMatch)
		g := &SearchPhotos{}
		require.NoError(t, Unserialize(g, "label:\"cat"))
		assert.Equal(t, "cat", g.Label)
	})
}

func TestParseQueryString_NoMatch(t *testing.T) {
	t.Run("NothingLeft", func(t *testing.T) {
		for name, f := range map[string]SearchForm{
			"Path":           &SearchPhotos{Path: "\xff"},
			"Name":           &SearchPhotos{Name: "\xff"},
			"Filename":       &SearchPhotos{Filename: "\x00\xfe"},
			"Scope":          &SearchPhotos{Scope: "\xff"},
			"Query":          &SearchPhotos{Query: "\x00"},
			"QueryValue":     &SearchPhotos{Query: "path:\xff label:cat"},
			"QuotedValue":    &SearchPhotos{Query: "path:\"\xff\x00\""},
			"TrimmedValue":   &SearchPhotos{Query: "path:\xff|"},
			"GeoPath":        &SearchPhotosGeo{Path: "\xff"},
			"GeoQueryValue":  &SearchPhotosGeo{Query: "name:\xff"},
			"AlbumsCategory": &SearchAlbums{Category: "\x00"},
			"LabelsQuery":    &SearchLabels{Query: "\xff"},
			"SubjectsName":   &SearchSubjects{Name: "\xff"},
			"FacesSubject":   &SearchFaces{Subject: "\xff"},
			"UsersName":      &SearchUsers{Name: "\xff"},
			"CamerasQuery":   &SearchCameras{Query: "\xff"},
			"QuerySpace":     &SearchPhotos{Query: "\x00 "},
			"QueryPipe":      &SearchPhotos{Query: "\xff|"},
			"PathSpace":      &SearchPhotos{Path: "\xff "},
			"PathWildcard":   &SearchPhotos{Path: "\xff*"},
			"LabelSpace":     &SearchPhotos{Label: "\xff "},
			"Filter":         &SearchPhotos{Filter: "\xff"},
			"ValueWildcard":  &SearchPhotos{Query: "label:\xff*"},
			"Before":         &SearchPhotos{Query: "before:\xff"},
			"After":          &SearchPhotos{Query: "after:\x00"},
			"Quality":        &SearchPhotos{Query: "quality:\xff"},
			"Public":         &SearchPhotos{Query: "public:\xff"},
			"GeoBefore":      &SearchPhotosGeo{Query: "before:\xff"},
		} {
			assert.ErrorIs(t, ParseQueryString(f), ErrNoMatch, name)
		}
	})
	t.Run("StoredFilter", func(t *testing.T) {
		// A stored folder album filter may name a folder that is not valid UTF-8, and its value is dropped.
		f := &SearchPhotos{Album: "as6sg6bxpogaaba8", Filter: "path:\xc4\xd6\xdc public:true"}
		require.NoError(t, f.ParseQueryString())
		assert.Equal(t, "", f.Path)
		assert.True(t, f.Public)
		g := &SearchPhotosGeo{Filter: "name:\"\x00\" public:true"}
		require.NoError(t, g.ParseQueryString())
		assert.True(t, g.Public)
	})
	t.Run("Methods", func(t *testing.T) {
		f := &SearchPhotos{Path: "\xff"}
		assert.ErrorIs(t, f.ParseQueryString(), ErrNoMatch)
		g := &SearchPhotosGeo{Name: "\xff"}
		assert.ErrorIs(t, g.ParseQueryString(), ErrNoMatch)
	})
	t.Run("Unchanged", func(t *testing.T) {
		for _, f := range []*SearchPhotos{
			{}, {Query: "path:"}, {Path: "2020/Holiday"}, {Path: "2020/"}, {Path: "20*"}, {Path: "2020|2021"},
			{Order: "\xff"}, {Filter: "path:2020"}, {Query: "label:c\xff path:"}, {Query: "label:\"a\xff b\" path:"},
			{Query: "pa\xffth:"}, {Query: "\x00path:"}, {Query: "path:|"}, {Path: "*"}, {Query: "|"}, {Query: "label:*"},
			{Query: "before:2020-01-01"}, {Query: "before:2020-01-01\xff"}, {Query: "label:\xff-"},
		} {
			p := *f
			require.NoError(t, f.ParseQueryString(), "%+v", p)
		}
	})
	t.Run("ValidCharactersLeft", func(t *testing.T) {
		f := &SearchPhotos{Query: "cat\xff"}
		require.NoError(t, f.ParseQueryString())
		assert.Equal(t, "cat", f.Query)
		f = &SearchPhotos{Query: "cat \xff"}
		require.NoError(t, f.ParseQueryString())
		assert.Equal(t, "cat", f.Query)
		f = &SearchPhotos{Query: "path:20\xff20"}
		require.NoError(t, f.ParseQueryString())
		assert.Equal(t, "2020", f.Path)
		f = &SearchPhotos{Path: "a|\x00"}
		require.NoError(t, f.ParseQueryString())
		assert.Equal(t, "a|", f.Path)
	})
	t.Run("InvalidDate", func(t *testing.T) {
		// A date with other characters left fails as an invalid date, not as a value that matches nothing.
		err := (&SearchPhotos{Query: "before:\xffx"}).ParseQueryString()
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrNoMatch)
	})
	t.Run("ReplacementCharacter", func(t *testing.T) {
		// A valid U+FFFD in a value is searched as it is.
		f := &SearchPhotos{Query: "path:\"\uFFFD\""}
		require.NoError(t, f.ParseQueryString())
		assert.Equal(t, "\uFFFD", f.Path)
	})
	t.Run("Unserialize", func(t *testing.T) {
		// Stored album filters keep dropping such values.
		f := &SearchPhotos{}
		require.NoError(t, Unserialize(f, "path:\xff label:cat"))
		assert.Equal(t, "", f.Path)
		assert.Equal(t, "cat", f.Label)
	})
}

func TestBindableStrings(t *testing.T) {
	t.Run("Struct", func(t *testing.T) {
		f := &SearchAlbums{Query: "a\x00b", Category: "\x00", Location: "Berlin"}
		assert.False(t, bindableStrings(f))
		assert.Equal(t, "ab", f.Query)
		assert.Equal(t, "", f.Category)
		assert.Equal(t, "Berlin", f.Location)
	})
	t.Run("Order", func(t *testing.T) {
		f := &SearchAlbums{Query: "a\x00b", Order: "\xff"}
		assert.True(t, bindableStrings(f))
		assert.Equal(t, "", f.Order)
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
