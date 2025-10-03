package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosFilterResolution(t *testing.T) {
	t.Run("Two", func(t *testing.T) {
		var f form.SearchPhotos

		f.Mp = "2"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 2, r.PhotoResolution)
			assert.LessOrEqual(t, 2, r.PhotoResolution)
		}
		assert.Len(t, photos, 8)
	})
	t.Run("OneNum50", func(t *testing.T) {
		var f form.SearchPhotos

		f.Mp = "1-50"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 50, r.PhotoResolution)
			assert.LessOrEqual(t, 1, r.PhotoResolution)
		}

		assert.Len(t, photos, 10)
	})
	t.Run("ThreeNum150", func(t *testing.T) {
		var f form.SearchPhotos

		f.Mp = "3-150"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 150, r.PhotoResolution)
			assert.LessOrEqual(t, 3, r.PhotoResolution)
		}

		assert.Len(t, photos, 3)
	})
	t.Run("Num155", func(t *testing.T) {
		var f form.SearchPhotos

		f.Mp = "155"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 0)
	})
	t.Run("Invalid", func(t *testing.T) {
		var f form.SearchPhotos

		f.Mp = "%gold"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)
	})
}

func TestPhotosQueryResolution(t *testing.T) {
	t.Run("Two", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "mp:\"2\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 2, r.PhotoResolution)
			assert.LessOrEqual(t, 2, r.PhotoResolution)
		}

		assert.Len(t, photos, 8)
	})
	t.Run("OneNum50", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "mp:\"1-50\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 50, r.PhotoResolution)
			assert.LessOrEqual(t, 1, r.PhotoResolution)
		}

		assert.Len(t, photos, 10)
	})
	t.Run("ThreeNum150", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "mp:\"3-150\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 150, r.PhotoResolution)
			assert.LessOrEqual(t, 3, r.PhotoResolution)
		}

		assert.Len(t, photos, 3)
	})
	t.Run("Eighteen", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "mp:\"18\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 0)
	})
	t.Run("Invalid", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "mp:\"%gold\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)
	})
}
