package form

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPhotoSearchForm(t *testing.T) {
	form := &PhotoSearch{}

	assert.IsType(t, new(PhotoSearch), form)
}

func TestParseQueryString(t *testing.T) {
	t.Run("path", func(t *testing.T) {
		form := &PhotoSearch{Query: "path:123abc/,EFG"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "123abc/,EFG", form.Path)
	})

	t.Run("folder", func(t *testing.T) {
		form := &PhotoSearch{Query: "folder:123abc/,EFG"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "123abc/,EFG", form.Path)
	})
	t.Run("valid query", func(t *testing.T) {
		form := &PhotoSearch{Query: "label:cat query:\"fooBar baz\" before:2019-01-15 camera:23 favorite:false dist:25000 lat:33.45343166666667"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "cat", form.Label)
		assert.Equal(t, "fooBar baz", form.Query)
		assert.Equal(t, 23, form.Camera)
		assert.Equal(t, time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), form.Before)
		assert.Equal(t, false, form.Favorite)
		assert.Equal(t, uint(0x61a8), form.Dist)
		assert.Equal(t, float32(33.45343), form.Lat)
	})
	t.Run("valid query 2", func(t *testing.T) {
		form := &PhotoSearch{Query: "chroma:200 title:\"test\" after:2018-01-15 duplicate:false  favorite:true lng:33.45343166666667"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, uint8(200), form.Chroma)
		assert.Equal(t, "test", form.Title)
		assert.Equal(t, time.Date(2018, 01, 15, 0, 0, 0, 0, time.UTC), form.After)
		assert.Equal(t, false, form.Duplicate)
		assert.Equal(t, float32(33.45343), form.Lng)
	})
	t.Run("valid query with filter", func(t *testing.T) {
		form := &PhotoSearch{Query: "label:cat title:\"fooBar baz\"", Filter: "label:dog"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "dog", form.Label)
		assert.Equal(t, "label:dog", form.Filter)
		assert.Equal(t, "fooBar baz", form.Title)
	})
	t.Run("valid query with umlauts", func(t *testing.T) {
		form := &PhotoSearch{Query: "title:\"tübingen\""}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "tübingen", form.Title)
	})
	t.Run("query for invalid filter", func(t *testing.T) {
		form := &PhotoSearch{Query: "xxx:false"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: Xxx", err.Error())
	})
	t.Run("query for favorites with uncommon bool value", func(t *testing.T) {
		form := &PhotoSearch{Query: "favorite:cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Favorite)
	})
	t.Run("query for lat with invalid type", func(t *testing.T) {
		form := &PhotoSearch{Query: "lat:cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		log.Debugf("%+v\n", form)

		assert.Equal(t, "strconv.ParseFloat: parsing \"cat\": invalid syntax", err.Error())
	})
	t.Run("query for dist with invalid type", func(t *testing.T) {
		form := &PhotoSearch{Query: "dist:cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		log.Debugf("%+v\n", form)

		assert.Equal(t, "strconv.Atoi: parsing \"cat\": invalid syntax", err.Error())
	})
	t.Run("query for camera with invalid type", func(t *testing.T) {
		form := &PhotoSearch{Query: "camera:cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		log.Debugf("%+v\n", form)

		assert.Equal(t, "strconv.Atoi: parsing \"cat\": invalid syntax", err.Error())
	})
	t.Run("query for before with invalid type", func(t *testing.T) {
		form := &PhotoSearch{Query: "before:cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		log.Debugf("%+v\n", form)

		assert.Equal(t, "Could not find format for \"cat\"", err.Error())
	})
}

func TestNewPhotoSearch(t *testing.T) {
	r := NewPhotoSearch("cat")
	assert.IsType(t, PhotoSearch{}, r)
}

func TestPhotoSearch_Serialize(t *testing.T) {
	form := PhotoSearch{
		Query:   "foo BAR",
		Private: true,
		Photo:   false,
		Lat:     1.5,
		Lng:     -10.33333,
		Year:    2002,
		Chroma:  1,
		Diff:    424242,
		Before:  time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC),
	}

	result := form.Serialize()

	t.Logf("SERIALIZED: %s", result)

	assert.IsType(t, "string", result)
}

func TestPhotoSearch_SerializeAll(t *testing.T) {
	form := PhotoSearch{
		Query:   "foo BAR",
		Private: true,
		Photo:   false,
		Lat:     1.5,
		Lng:     -10.33333,
		Year:    2002,
		Chroma:  1,
		Diff:    424242,
		Before:  time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC),
	}

	result := form.SerializeAll()

	t.Logf("SERIALIZED: %s", result)

	assert.IsType(t, "string", result)
}
