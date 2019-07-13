package photoprism

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/forms"
)

func TestSearch_Photos_Query(t *testing.T) {
	conf := config.TestConfig()

	conf.CreateDirectories()

	search := NewSearch(conf.OriginalsPath(), conf.Db())

	t.Run("normal query", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "animal"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos[0])
	})
	t.Run("label query", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "label:dog"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("invalid label query", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "label:xxx"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		assert.Equal(t, err.Error(), "label \"xxx\" not found")

		t.Log(photos)
	})
	t.Run("form.location true", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = ""
		form.Count = 3
		form.Offset = 0
		form.Location = true

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.camera", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = ""
		form.Count = 3
		form.Offset = 0
		form.Camera = 2

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.color", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = ""
		form.Count = 3
		form.Offset = 0
		form.Color = "blue"

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.favorites", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "favorites:true"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.country", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "country:de"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.title", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "title:Pug Dog"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.description", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "description:xxx"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.notes", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "notes:xxx"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.hash", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "hash:xxx"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.duplicate", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "duplicate:true"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.portrait", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "portrait:true"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.mono", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "mono:true"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.chroma", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "chroma:50"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.fmin and Order:oldest", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "Fmin:5 Order:oldest"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.fmax and Order:newest", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "Fmax:2 Order:newest"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.Lat and form.Long and Order:imported", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "Lat:33.45343166666667 Long:25.764711666666667 Dist:2000 Order:imported"
		form.Count = 3
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})
	t.Run("form.Before and form.After", func(t *testing.T) {
		var form forms.PhotoSearchForm
		form.Query = "Before:2005-01-01 After:2003-01-01"
		form.Count = 5000
		form.Offset = 0

		photos, err := search.Photos(form)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(photos)
	})

}
