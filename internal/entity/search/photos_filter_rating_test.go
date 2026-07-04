package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosQueryRating(t *testing.T) {
	var f0 form.SearchPhotos

	f0.Merged = true

	// Parse query string and filter.
	if err := f0.ParseQueryString(); err != nil {
		t.Fatal(err)
	}

	photos0, _, err := Photos(f0)

	if err != nil {
		t.Fatal(err)
	}

	all := len(photos0)

	t.Run("MinimumOneStar", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "rating:1"
		f.Merged = true

		if err := f.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, f.Rating)

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		// The default fixtures carry no explicit star ratings.
		assert.Len(t, photos, 0)
	})
	t.Run("HigherMinimumIsSubset", func(t *testing.T) {
		var f1 form.SearchPhotos

		f1.Query = "rating:1"
		f1.Merged = true

		if err := f1.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos1, _, err1 := Photos(f1)

		if err1 != nil {
			t.Fatal(err1)
		}

		var f5 form.SearchPhotos

		f5.Query = "rating:5"
		f5.Merged = true

		if err := f5.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos5, _, err5 := Photos(f5)

		if err5 != nil {
			t.Fatal(err5)
		}

		assert.LessOrEqual(t, len(photos5), len(photos1))
		assert.LessOrEqual(t, len(photos1), all)
	})
	t.Run("ZeroMeansNoFilter", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "rating:0"
		f.Merged = true

		if err := f.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, all)
	})
}
