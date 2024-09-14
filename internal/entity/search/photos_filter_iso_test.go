package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestPhotosFilterIso(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		var f form.SearchPhotos

		f.Iso = "200"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 200, r.PhotoIso)
			assert.LessOrEqual(t, 200, r.PhotoIso)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("200-400", func(t *testing.T) {
		var f form.SearchPhotos

		f.Iso = "200-400"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 400, r.PhotoIso)
			assert.LessOrEqual(t, 200, r.PhotoIso)
		}

		assert.Equal(t, len(photos), 4)
	})
	t.Run("1-400", func(t *testing.T) {
		var f form.SearchPhotos

		f.Iso = "1-400"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 400, r.PhotoIso)
			assert.LessOrEqual(t, 1, r.PhotoIso)
		}

		assert.Equal(t, len(photos), 5)
	})
	t.Run("155", func(t *testing.T) {
		var f form.SearchPhotos

		f.Iso = "155"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("-100", func(t *testing.T) {
		var f form.SearchPhotos

		f.Iso = "-100"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 40)
	})
	t.Run("invalid", func(t *testing.T) {
		var f form.SearchPhotos

		f.Iso = "%gold"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)
	})
}

func TestPhotosQueryIso(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "iso:\"200\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 200, r.PhotoIso)
			assert.LessOrEqual(t, 200, r.PhotoIso)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("200-400", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "iso:\"200-400\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 400, r.PhotoIso)
			assert.LessOrEqual(t, 200, r.PhotoIso)
		}

		assert.Equal(t, len(photos), 4)
	})
	t.Run("1-400", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "iso:\"1-400\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 400, r.PhotoIso)
			assert.LessOrEqual(t, 1, r.PhotoIso)
		}

		assert.Equal(t, len(photos), 5)
	})
	t.Run("155", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "iso:\"155\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("-100", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "iso:\"-100\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)
	})
	t.Run("invalid", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "iso:\"%gold\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)
	})
}
