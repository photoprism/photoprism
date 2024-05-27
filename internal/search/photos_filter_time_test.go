package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/sortby"
)

func TestPhotosFilterTime(t *testing.T) {
	t.Run("Added", func(t *testing.T) {
		var f form.SearchPhotos

		timeStamp, err := time.Parse(time.RFC3339, "2021-01-02T00:00:00Z")

		if err != nil {
			t.Fatal(err)
		}

		f.Added = timeStamp
		f.Order = sortby.Added
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("Added: %#v", photos)

		assert.GreaterOrEqual(t, 40, len(photos))
	})
	t.Run("Updated", func(t *testing.T) {
		var f form.SearchPhotos

		timeStamp, err := time.Parse(time.RFC3339, "2022-01-02T13:04:05-01:00")

		if err != nil {
			t.Fatal(err)
		}

		f.Updated = timeStamp
		f.Order = sortby.Updated
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("Updated: %#v", photos)

		assert.GreaterOrEqual(t, 49, len(photos))
	})
	t.Run("Edited", func(t *testing.T) {
		var f form.SearchPhotos

		timeStamp, err := time.Parse(time.RFC3339, "2020-01-01T12:00:00Z")

		if err != nil {
			t.Fatal(err)
		}

		f.Edited = timeStamp
		f.Order = sortby.Edited
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("Edited: %#v", photos)

		assert.GreaterOrEqual(t, 2, len(photos))
	})
	t.Run("Taken", func(t *testing.T) {
		var f form.SearchPhotos

		timeStamp, err := time.Parse(time.RFC3339, "2014-07-17T15:42:12Z")

		if err != nil {
			t.Fatal(err)
		}

		f.Taken = timeStamp
		f.Order = sortby.Added
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("Taken: %#v", photos)

		assert.GreaterOrEqual(t, 1, len(photos))
	})
	t.Run("Before", func(t *testing.T) {
		var f form.SearchPhotos

		timeStamp, err := time.Parse(time.RFC3339, "2022-01-02T13:04:05Z")

		if err != nil {
			t.Fatal(err)
		}

		f.Before = timeStamp
		f.Order = sortby.Added
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("Before: %#v", photos)

		assert.GreaterOrEqual(t, 47, len(photos))
	})
	t.Run("After", func(t *testing.T) {
		var f form.SearchPhotos

		timeStamp, err := time.Parse(time.RFC3339, "2022-01-02T13:04:05Z")

		if err != nil {
			t.Fatal(err)
		}

		f.After = timeStamp
		f.Order = sortby.Added
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("After: %#v", photos)

		assert.GreaterOrEqual(t, 2, len(photos))
	})
}
