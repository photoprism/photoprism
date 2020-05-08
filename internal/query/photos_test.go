package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotos(t *testing.T) {
	t.Run("normal query", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = ""
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("label query", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "label:dog"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			// TODO: Add database fixtures to avoid failing queries
			t.Logf("query failed: %s", err.Error())
			// t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("invalid label query", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "label:xxx"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		assert.Error(t, err)
		assert.Empty(t, photos)

		if err != nil {
			assert.Equal(t, err.Error(), "label xxx not found")
		}
	})
	t.Run("form.location true", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = ""
		f.Count = 3
		f.Offset = 0
		f.Location = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.camera", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = ""
		f.Count = 3
		f.Offset = 0
		f.Camera = 2

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.color", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = ""
		f.Count = 3
		f.Offset = 0
		f.Color = "blue"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.favorites", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "favorites:true"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.country", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "country:de"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.title", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "title:Pug Dog"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.hash", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "hash:xxx"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.duplicate", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "duplicate:true"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.portrait", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "portrait:true"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.mono", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "mono:true"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.chroma", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "chroma:50"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.fmin and Order:oldest", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "Fmin:5 Order:oldest"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.fmax and Order:newest", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "Fmax:2 Order:newest"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.Lat and form.Lng and Order:imported", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "Lat:33.45343166666667 Lng:25.764711666666667 Dist:2000 Order:imported"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})
	t.Run("form.Before and form.After", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "Before:2005-01-01 After:2003-01-01"
		f.Count = 5000
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
	})

}
