package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestPhotosFilterFNumber(t *testing.T) {
	t.Run("3.2", func(t *testing.T) {
		var f form.SearchPhotos

		f.F = "3.2"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, float32(3.2), r.PhotoFNumber)
			assert.LessOrEqual(t, float32(3.2), r.PhotoFNumber)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("3.5-5", func(t *testing.T) {
		var f form.SearchPhotos

		f.F = "3.5-5"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, float32(5), r.PhotoFNumber)
			assert.LessOrEqual(t, float32(3.5), r.PhotoFNumber)
		}

		assert.Equal(t, len(photos), 3)
	})
	t.Run("3-10", func(t *testing.T) {
		var f form.SearchPhotos

		f.F = "3-10"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, float32(10), r.PhotoFNumber)
			assert.LessOrEqual(t, float32(3), r.PhotoFNumber)
		}

		assert.Equal(t, len(photos), 5)
	})
	t.Run("8", func(t *testing.T) {
		var f form.SearchPhotos

		f.F = "8"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("-100", func(t *testing.T) {
		var f form.SearchPhotos

		f.F = "-100"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 40)
	})
	t.Run("invalid", func(t *testing.T) {
		var f form.SearchPhotos

		f.F = "%gold"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)
	})
}

func TestPhotosQueryFNumber(t *testing.T) {
	t.Run("3.2", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "f:\"3.2\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, float32(3.2), r.PhotoFNumber)
			assert.LessOrEqual(t, float32(3.2), r.PhotoFNumber)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("3.5-5", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "f:\"3.5-5\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, float32(5), r.PhotoFNumber)
			assert.LessOrEqual(t, float32(3.5), r.PhotoFNumber)
		}
		assert.Equal(t, len(photos), 3)
	})
	t.Run("3-10", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "f:\"3-10\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, float32(10), r.PhotoFNumber)
			assert.LessOrEqual(t, float32(3), r.PhotoFNumber)
		}

		assert.Equal(t, len(photos), 5)
	})
	t.Run("8", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "f:\"8\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("-100", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "f:\"-100\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)
	})
	t.Run("invalid", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "f:\"%gold\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)

	})
}
