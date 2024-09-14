package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestPhotosFilterAlt(t *testing.T) {
	t.Run("-10", func(t *testing.T) {
		var f form.SearchPhotos

		f.Alt = "-10"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, -10, r.PhotoAltitude)
			assert.LessOrEqual(t, -10, r.PhotoAltitude)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("-100--5", func(t *testing.T) {
		var f form.SearchPhotos

		f.Alt = "-100--5"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, -5, r.PhotoAltitude)
			assert.LessOrEqual(t, -100, r.PhotoAltitude)
		}

		assert.Equal(t, len(photos), 2)
	})
	t.Run("200-500", func(t *testing.T) {
		var f form.SearchPhotos

		f.Alt = "200-500"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 500, r.PhotoAltitude)
			assert.LessOrEqual(t, 200, r.PhotoAltitude)
		}

		assert.Equal(t, len(photos), 2)
	})
	t.Run("200", func(t *testing.T) {
		var f form.SearchPhotos

		f.Alt = "200"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("invalid", func(t *testing.T) {
		var f form.SearchPhotos

		f.Alt = "%gold"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)
	})
}

func TestPhotosQueryAlt(t *testing.T) {
	t.Run("-10", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "alt:\"-10\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, -10, r.PhotoAltitude)
			assert.LessOrEqual(t, -10, r.PhotoAltitude)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("-100--5", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "alt:\"-100--5\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, -5, r.PhotoAltitude)
			assert.LessOrEqual(t, -100, r.PhotoAltitude)
		}

		assert.Equal(t, len(photos), 2)
	})
	t.Run("200-500", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "alt:\"200-500\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 500, r.PhotoAltitude)
			assert.LessOrEqual(t, 200, r.PhotoAltitude)
		}

		assert.Equal(t, len(photos), 2)
	})
	t.Run("200", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "alt:\"200\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("invalid", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "alt:\"%gold\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)
	})
}
