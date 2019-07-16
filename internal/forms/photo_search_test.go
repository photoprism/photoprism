package forms

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
)

func TestPhotoSearchForm(t *testing.T) {
	form := &PhotoSearchForm{}

	assert.IsType(t, new(PhotoSearchForm), form)
}

func TestParseQueryString(t *testing.T) {

	t.Run("valid query", func(t *testing.T) {
		form := &PhotoSearchForm{Query: "label:cat query:\"fooBar baz\" before:2019-01-15 camera:23 favorites:false dist:25000 lat:33.45343166666667"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Nil(t, err)
		assert.Equal(t, "cat", form.Label)
		assert.Equal(t, "foobar baz", form.Query)
		assert.Equal(t, 23, form.Camera)
		assert.Equal(t, time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), form.Before)
		assert.Equal(t, false, form.Favorites)
		assert.Equal(t, uint(0x61a8), form.Dist)
		assert.Equal(t, 33.45343166666667, form.Lat)
	})
	t.Run("valid query 2", func(t *testing.T) {
		form := &PhotoSearchForm{Query: "chroma:600 description:\"test\" after:2018-01-15 duplicate:false  favorites:true long:33.45343166666667"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Nil(t, err)
		assert.Equal(t, uint(0x258), form.Chroma)
		assert.Equal(t, "test", form.Description)
		assert.Equal(t, time.Date(2018, 01, 15, 0, 0, 0, 0, time.UTC), form.After)
		assert.Equal(t, false, form.Duplicate)
		assert.Equal(t, 33.45343166666667, form.Long)
	})
	t.Run("query for invalid filter", func(t *testing.T) {
		form := &PhotoSearchForm{Query: "xxx:false"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: Xxx", err.Error())
	})
	t.Run("query for favorites with invalid type", func(t *testing.T) {
		form := &PhotoSearchForm{Query: "favorites:cat"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Equal(t, "not a bool value: Favorites", err.Error())
	})
	t.Run("query for lat with invalid type", func(t *testing.T) {
		form := &PhotoSearchForm{Query: "lat:cat"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Equal(t, "strconv.ParseFloat: parsing \"cat\": invalid syntax", err.Error())
	})
	t.Run("query for dist with invalid type", func(t *testing.T) {
		form := &PhotoSearchForm{Query: "dist:cat"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Equal(t, "strconv.Atoi: parsing \"cat\": invalid syntax", err.Error())
	})
	t.Run("query for camera with invalid type", func(t *testing.T) {
		form := &PhotoSearchForm{Query: "camera:cat"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Equal(t, "strconv.Atoi: parsing \"cat\": invalid syntax", err.Error())
	})
	t.Run("query for before with invalid type", func(t *testing.T) {
		form := &PhotoSearchForm{Query: "before:cat"}

		err := form.ParseQueryString()

		log.Debugf("%+v\n", form)

		assert.Equal(t, "Could not find format for \"cat\"", err.Error())
	})
}
