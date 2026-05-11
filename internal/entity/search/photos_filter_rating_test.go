package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosQueryRating(t *testing.T) {
	t.Run("Exact", func(t *testing.T) {
		f := form.NewSearchPhotos("rating:5")

		if err := f.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		requireContainsPhoto(t, photos, "Photo01")
		for _, photo := range photos {
			assert.Equal(t, 5, photo.PhotoRating)
		}
	})

	t.Run("Range", func(t *testing.T) {
		f := form.NewSearchPhotos("rating:4-5")

		if err := f.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		requireContainsPhoto(t, photos, "Photo01")
		requireContainsPhoto(t, photos, "bridge2")
		for _, photo := range photos {
			assert.GreaterOrEqual(t, photo.PhotoRating, 4)
			assert.LessOrEqual(t, photo.PhotoRating, 5)
		}
	})
}

// requireContainsPhoto fails the test if the result set does not include a photo name.
func requireContainsPhoto(t *testing.T, photos PhotoResults, name string) {
	t.Helper()

	for _, photo := range photos {
		if photo.PhotoName == name {
			return
		}
	}

	t.Fatalf("expected photo %q in results", name)
}
