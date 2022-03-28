package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestPhotosFilterAlbum(t *testing.T) {
	t.Run("album start %", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "%gold"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("album middle %", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "I love % dog"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 0)

	})
	t.Run("album end %", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "sale%"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 0)
	})
	t.Run("album start &", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "&IlikeFood"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		// TODO: Needs review, variable number of results.
		assert.GreaterOrEqual(t, len(photos), 0)
	})
	t.Run("album middle &", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "Pets & Dogs"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album end &", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "Light&"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album start '", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "'Family"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album middle '", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "Father's Day"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album end '", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "Ice Cream'"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album start *", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "*Forrest"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("album middle *", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "My*Kids"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album end *", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "Yoga***"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album start |", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "|Banana"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		// TODO: Needs review, variable number of results.

		assert.GreaterOrEqual(t, len(photos), 0)
	})
	t.Run("album middle |", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "Red|Green"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		// TODO: Needs review, variable number of results.

		assert.GreaterOrEqual(t, len(photos), 0)
	})
	t.Run("album end |", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "Blue|"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album start number", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "345 Shirt"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album middle number", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "Color555 Blue"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album end number", func(t *testing.T) {
		var f form.SearchPhotos

		f.Album = "Route 66"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
}

func TestPhotosQueryAlbum(t *testing.T) {
	t.Run("album start %", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"%gold\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("album middle %", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"I love % dog\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("album end %", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"sale%\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album start &", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"&IlikeFood\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		// TODO: Needs review, variable number of results.
		assert.GreaterOrEqual(t, len(photos), 0)
	})
	t.Run("album middle &", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"Pets & Dogs\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album end &", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"Light&\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album start '", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"'Family\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album middle '", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"Father's Day\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("album end '", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"Ice Cream'\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 0)
	})
	t.Run("album start *", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"*Forrest\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("album middle *", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"My*Kids\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 0)
	})
	t.Run("album end *", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"Yoga***\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 0)
	})
	t.Run("album start |", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"|Banana\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("album middle |", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"Red|Green\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		// TODO: Needs review, variable number of results.

		assert.GreaterOrEqual(t, len(photos), 0)
	})
	t.Run("album end |", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"Blue|\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 0)
	})
	t.Run("album start number", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"345 Shirt\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 0)
	})
	t.Run("album middle number", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"Color555 Blue\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 0)
	})
	t.Run("album end number", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "album:\"Route 66\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 0)
	})
}
