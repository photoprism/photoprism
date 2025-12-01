package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosQueryPortrait(t *testing.T) {
	portraitSearchForm := &form.SearchPhotos{
		Query:  "portrait:true",
		Merged: true,
	}

	// Parse query string and filter.
	if err := portraitSearchForm.ParseQueryString(); err != nil {
		t.Fatal(err)
	}

	portraits, _, findErr := Photos(*portraitSearchForm)

	if findErr != nil {
		t.Fatal(findErr)
	}

	assert.GreaterOrEqual(t, len(portraits), 39)

	t.Run("FalseGreaterThanYes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:yes"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))

		f.Query = "portrait:false"
		f.Merged = true

		allPhotos, _, err2 := Photos(f)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Greater(t, len(allPhotos), len(photos))
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"%gold\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"I love % dog\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"sale%\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"&IlikeFood\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"Pets & Dogs\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, portraits, len(photos))
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"Light&\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"'Family\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, portraits, len(photos))
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		// Note: If the string in portrait starts with f/F, the txt package will assume it means false,
		f.Query = "portrait:\"Mother's Day\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"Ice Cream'\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"*Forrest\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"My*Kids\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"Yoga***\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"|Banana\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"Red|Green\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"Blue|\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"345 Shirt\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"Color555 Blue\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"Route 66\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("AndSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"Route 66 & Father's Day\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("OrSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "portrait:\"Route %66 | *Father's Day\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Len(t, portraits, len(photos))
	})
	t.Run("Landscape", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "landscape:true"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Equal(t, 8, len(photos))
	})
	t.Run("Square", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "square:true"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(photos))
	})
}
