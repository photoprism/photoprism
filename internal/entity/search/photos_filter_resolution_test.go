package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestPhotosFilterResolution(t *testing.T) {
	t.Run("2", func(t *testing.T) {
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
		assert.Equal(t, len(photos), 8)
	})
	t.Run("1-50", func(t *testing.T) {
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

		assert.Equal(t, len(photos), 9)
	})
	t.Run("3-150", func(t *testing.T) {
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

		assert.Equal(t, len(photos), 2)
	})
	t.Run("155", func(t *testing.T) {
		var f form.SearchPhotos

		f.Mp = "155"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("invalid", func(t *testing.T) {
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
	t.Run("2", func(t *testing.T) {
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

		assert.Equal(t, len(photos), 8)
	})
	t.Run("1-50", func(t *testing.T) {
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

		assert.Equal(t, len(photos), 9)
	})
	t.Run("3-150", func(t *testing.T) {
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

		assert.Equal(t, len(photos), 2)
	})
	t.Run("18", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "mp:\"18\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("invalid", func(t *testing.T) {
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
