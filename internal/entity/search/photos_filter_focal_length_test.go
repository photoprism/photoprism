package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosFilterFocalLength(t *testing.T) {
	t.Run("Num28", func(t *testing.T) {
		var f form.SearchPhotos

		f.Mm = "28"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 28, r.PhotoFocalLength)
			assert.LessOrEqual(t, 28, r.PhotoFocalLength)
		}
		assert.Len(t, photos, 1)
	})
	t.Run("Num28Num50", func(t *testing.T) {
		var f form.SearchPhotos

		f.Mm = "28-50"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 50, r.PhotoFocalLength)
			assert.LessOrEqual(t, 28, r.PhotoFocalLength)
		}

		assert.Len(t, photos, 4)
	})
	t.Run("OneNum400", func(t *testing.T) {
		var f form.SearchPhotos

		f.Mm = "1-400"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 400, r.PhotoFocalLength)
			assert.LessOrEqual(t, 1, r.PhotoFocalLength)
		}

		assert.Len(t, photos, 6)
	})
	t.Run("Num22", func(t *testing.T) {
		var f form.SearchPhotos

		f.Mm = "22"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 0)
	})
	t.Run("Num100", func(t *testing.T) {
		var f form.SearchPhotos

		f.Mm = "-100"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 40)
	})
	t.Run("Invalid", func(t *testing.T) {
		var f form.SearchPhotos

		f.Mm = "%gold"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)
	})
}

func TestPhotosQueryFocalLength(t *testing.T) {
	t.Run("Num28", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "mm:\"28\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 28, r.PhotoFocalLength)
			assert.LessOrEqual(t, 28, r.PhotoFocalLength)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("Num28Num30", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "mm:\"28-30\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 30, r.PhotoFocalLength)
			assert.LessOrEqual(t, 28, r.PhotoFocalLength)
		}

		assert.Len(t, photos, 2)
	})
	t.Run("OneNum400", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "mm:\"1-400\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, r := range photos {
			assert.GreaterOrEqual(t, 400, r.PhotoFocalLength)
			assert.LessOrEqual(t, 1, r.PhotoFocalLength)
		}

		assert.Len(t, photos, 6)
	})
	t.Run("Eighteen", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "mm:\"18\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 0)
	})
	t.Run("Num100", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "mm:\"-100\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)
	})
	t.Run("Invalid", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "mm:\"%gold\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.GreaterOrEqual(t, len(photos), 40)
	})
}
