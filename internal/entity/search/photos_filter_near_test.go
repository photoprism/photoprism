package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosFilterNear(t *testing.T) {
	t.Run("ps6sg6be2lvl0y24", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "ps6sg6be2lvl0y24"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 13, len(photos))
	})
	t.Run("ps6sg6byk7wrbk30", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "ps6sg6byk7wrbk30"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 29, len(photos))
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "%gold"
		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "I love % dog"
		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "sale%"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "&IlikeFood"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "Pets & Dogs"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "Light&"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "'Family"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "Father's type"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "Ice Cream'"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "*Forrest"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "My*Kids"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "Yoga***"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "|Banana"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "Red|Green"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "Blue|"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "345 Shirt"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "type555 Blue"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Near = "Route 66"

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
}

func TestPhotosQueryNear(t *testing.T) {
	t.Run("ps6sg6be2lvl0y24", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:ps6sg6be2lvl0y24"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 13, len(photos))
	})
	t.Run("ps6sg6byk7wrbk30", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:ps6sg6byk7wrbk30"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 29, len(photos))
	})
	t.Run("ps6sg6be2lvl0y24 pipe ps6sg6byk7wrbk30", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:ps6sg6be2lvl0y24|ps6sg6byk7wrbk30"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 42, len(photos))
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"%gold\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"I love % dog\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"sale%\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"&IlikeFood\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"Pets & Dogs\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"Light&\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"'Family\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"Father's type\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"Ice Cream'\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"*Forrest\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"My*Kids\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"Yoga***\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"|Banana\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"Red|Green\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"Blue|\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"345 Shirt\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"type555 Blue\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "near:\"Route 66\""

		_, _, err := Photos(f)

		assert.Equal(t, err.Error(), "Not found")
	})
}
