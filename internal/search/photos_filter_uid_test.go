package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosFilterUid(t *testing.T) {
	t.Run("ps6sg6be2lvl0yh0", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "ps6sg6be2lvl0yh0"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("ps6sg6be2lvl0yh*", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "ps6sg6be2lvl0yh*"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "%gold"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "I love % dog"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "sale%"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "&IlikeFood"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "Pets & Dogs"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "Light&"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "'Family"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "Father's uid"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "Ice Cream'"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "*Forrest"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "My*Kids"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "Yoga***"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "|Banana"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "Red|Green"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "Blue|"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "345 Shirt"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "uid555 Blue"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.UID = "Route 66"
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
}

func TestPhotosQueryUid(t *testing.T) {
	t.Run("PhotoUID", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:ps6sg6be2lvl0yh0"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("QuotedPhotoUID", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"ps6sg6be2lvl0yh0\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("ps6sg6be2lvl0yh*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"ps6sg6be2lvl0yh*\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"%gold\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"I love % dog\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"sale%\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"&IlikeFood\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"Pets & Dogs\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"Light&\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"'Family\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"Father's uid\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"Ice Cream'\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"*Forrest\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"My*Kids\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"Yoga***\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"|Banana\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"Red|Green\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"Blue|\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"345 Shirt\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"uid555 Blue\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "uid:\"Route 66\""
		f.Merged = true

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Equal(t, len(photos), 0)
	})
}
