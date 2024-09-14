package form

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/stretchr/testify/assert"
)

func TestSearchPhotosGeo(t *testing.T) {
	t.Run("subjects", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "subjects:\"Jens Mander\""}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Jens Mander", form.Subjects)
	})
	t.Run("subject", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "subject:\"Jens\""}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Jens", form.Subject)
		assert.Equal(t, "", form.Person)
	})
	t.Run("id", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "id:\"ii3e4567-e89b-hdgtr\""}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "ii3e4567-e89b-hdgtr", form.ID)
	})
	t.Run("aliases", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "people:\"Jens & Mander\" folder:Foo person:Bar"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", form.Folder)
		assert.Equal(t, "", form.Person)
		assert.Equal(t, "", form.People)
		assert.Equal(t, "Foo", form.Path)
		assert.Equal(t, "Bar", form.Subject)
		assert.Equal(t, "Jens & Mander", form.Subjects)
	})
	t.Run("keywords", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "keywords:\"Foo Bar\""}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Foo Bar", form.Keywords)
	})
	t.Run("path", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "path:123abc/,EFG"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "123abc/,EFG", form.Path)
	})
	t.Run("name", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "name:filename.jpg"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "filename", form.Name)
	})
	t.Run("valid query", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "q:\"fooBar baz\" before:2019-01-15 dist:25000 lat:33.45343166666667"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal("err should be nil")
		}

		// log.Debugf("%+v\n", form)

		assert.Equal(t, "fooBar baz", form.Query)
		assert.Equal(t, time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), form.Before)
		assert.Equal(t, 25000.0, form.Dist)
		assert.Equal(t, 33.45343166666667, form.Lat)
	})
	t.Run("valid query path empty folder not empty", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "q:\"fooBar baz\" before:2019-01-15 dist:25000 lat:33.45343166666667 folder:test"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal("err should be nil")
		}

		// log.Debugf("%+v\n", form)

		assert.Equal(t, "fooBar baz", form.Query)
		assert.Equal(t, "test", form.Path)
		assert.Equal(t, "", form.Folder)
		assert.Equal(t, time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), form.Before)
		assert.Equal(t, 25000.0, form.Dist)
		assert.Equal(t, 33.45343166666667, form.Lat)
	})
	t.Run("valid query with filter", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "keywords:cat title:\"fooBar baz\"", Filter: "keywords:dog"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "dog", form.Keywords)
		assert.Equal(t, "keywords:dog", form.Filter)
		assert.Equal(t, "fooBar baz", form.Title)
	})
	t.Run("PortraitLandscapeSquare", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "portrait:true landscape:yes square:jo"}

		assert.False(t, form.Portrait)
		assert.False(t, form.Landscape)
		assert.False(t, form.Square)
		assert.False(t, form.Panorama)

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Portrait)
		assert.True(t, form.Landscape)
		assert.True(t, form.Square)
		assert.False(t, form.Panorama)
	})
	t.Run("AnimatedYes", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "animated:yes"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, form.Vector)
		assert.True(t, form.Animated)
	})
	t.Run("VectorYes", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "vector:yes"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, form.Animated)
		assert.True(t, form.Vector)
	})
	t.Run("query for favorites with uncommon bool value", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "favorite:cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "cat", form.Favorite)
	})
	t.Run("query for before with invalid type", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "before:cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		assert.Equal(t, "invalid before date", err.Error())
	})
	t.Run("query for lat with invalid type", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "lat:&cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		assert.Contains(t, err.Error(), "invalid syntax")
	})
	t.Run("query for quality with invalid type", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "quality:`cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		assert.Contains(t, err.Error(), "invalid syntax")
	})
	t.Run("query for dist with invalid type", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "dist:c@t"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		assert.Contains(t, err.Error(), "invalid syntax")
	})
	t.Run("Added", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "added:\"2022-01-02T13:04:05+01:00\""}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2022-01-02 13:04:05 +0100 UTC+01:00", form.Added.String())
		assert.Equal(t, "2022-01-02 12:04:05 +0000 UTC", form.Added.UTC().String())
		assert.Equal(t, "2022-01-02T13:04:05+01:00", form.Added.Format(time.RFC3339))
		assert.Equal(t, "2022-01-02T12:04:05Z", form.Added.UTC().Format(time.RFC3339))
	})
	t.Run("Updated", func(t *testing.T) {
		form := &SearchPhotosGeo{Query: "updated:\"2001-01-02 17:04:05\""}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2001-01-02 17:04:05 +0000 UTC", form.Updated.String())
		assert.Equal(t, "2001-01-02 17:04:05 +0000 UTC", form.Updated.UTC().String())
		assert.Equal(t, "2001-01-02T17:04:05Z", form.Updated.Format(time.RFC3339))
		assert.Equal(t, "2001-01-02T17:04:05Z", form.Updated.UTC().Format(time.RFC3339))
	})
	t.Run("MustBindWith", func(t *testing.T) {
		form := &SearchPhotosGeo{}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		u, err := url.Parse("https://www.photoprism.app/api/v1/photos?count=100&offset=0&order=added&added=2022-01-02T13:04:05-01:00&updated=2001-01-02T17:04:05Z&q=")

		if err != nil {
			t.Fatal(err)
		}

		c.Request = &http.Request{
			Header: make(http.Header),
			URL:    u,
		}

		// Abort if request params are invalid.
		if err = c.MustBindWith(form, binding.Form); err != nil {
			t.Fatal(err)
		}

		err = form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2022-01-02T13:04:05-01:00", form.Added.Format(time.RFC3339))
		assert.Equal(t, "2022-01-02T14:04:05Z", form.Added.UTC().Format(time.RFC3339))
		assert.Equal(t, "2001-01-02T17:04:05Z", form.Updated.Format(time.RFC3339))
		assert.Equal(t, 100, form.Count)
		assert.Equal(t, 0, form.Offset)
	})
}

func TestSearchPhotosGeo_Serialize(t *testing.T) {
	form := &SearchPhotosGeo{Query: "q:\"fooBar baz\"", Favorite: "true"}

	assert.Equal(t, "q:\"q:fooBar baz\" favorite:true", form.Serialize())
}

func TestSearchPhotosGeo_Unserialize(t *testing.T) {
	filter := "public:true label:bay|beach|cape|seashore"
	frm := SearchPhotosGeo{}
	err := Unserialize(&frm, filter)
	assert.Equal(t, true, frm.Public)
	assert.Equal(t, "bay|beach|cape|seashore", frm.Label)
	assert.NoError(t, err)
}

// public:true label:bay|beach|cape|seashore
func TestSearchPhotosGeo_SerializeAll(t *testing.T) {
	form := &SearchPhotosGeo{Query: "q:\"fooBar baz\"", Favorite: "true"}

	assert.Equal(t, "q:\"q:fooBar baz\" favorite:true", form.SerializeAll())
}

func TestNewSearchPhotosGeo(t *testing.T) {
	r := NewSearchPhotosGeo("Berlin")
	assert.IsType(t, SearchPhotosGeo{}, r)
}

func TestSearchPhotosGeo_FindUidOnly(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		f := &SearchPhotosGeo{UID: "priqwb43p5dh7777"}

		assert.True(t, f.FindUidOnly())
	})
	t.Run("false", func(t *testing.T) {
		f := &SearchPhotosGeo{Query: "label:cat", UID: "priqwb43p5dh7777"}

		assert.False(t, f.FindUidOnly())
	})
}
