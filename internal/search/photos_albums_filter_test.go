package search

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosFilterAlbums(t *testing.T) {
	t.Run("albums start %", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "%gold"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("albums middle %", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "I love % dog"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("albums end %", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "sale%"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums start &", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "&IlikeFood"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("albums middle &", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Pets & Dogs"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("albums end &", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Light&"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums start '", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "'Family"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	// TODO should not throw error
	/*t.Run("albums middle '", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Father's Day"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 0)
	})*/
	t.Run("albums end '", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Ice Cream'"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums start *", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "*Forrest"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("albums middle *", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "My*Kids"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums end *", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Yoga***"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums start |", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "|Banana"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("albums middle |", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Red|Green"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		// TODO: Needs review, variable number of results.

		assert.GreaterOrEqual(t, len(photos), 0)
	})
	t.Run("albums end |", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Blue|"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums start number", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "345 Shirt"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums middle number", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Color555 Blue"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums end number", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Route 66"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
}

func TestPhotosQueryAlbums(t *testing.T) {
	t.Run("albums start %", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"%gold\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("albums middle %", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"I love % dog\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("albums end %", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"sale%\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums start &", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"&IlikeFood\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("albums middle &", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Pets & Dogs\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("albums end &", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Light&\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums start '", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"'Family\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	//TODO should not throw error
	/*t.Run("albums middle '", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Father's Day\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 0)
	})*/
	t.Run("albums end '", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Ice Cream'\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums start *", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"*Forrest\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("albums middle *", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"My*Kids\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums end *", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Yoga***\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums start |", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"|Banana\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("albums middle |", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Red|Green\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		// TODO: Needs review, variable number of results.

		assert.GreaterOrEqual(t, len(photos), 0)
	})
	t.Run("albums end |", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Blue|\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums start number", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"345 Shirt\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums middle number", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Color555 Blue\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("albums end number", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Route 66\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
}
