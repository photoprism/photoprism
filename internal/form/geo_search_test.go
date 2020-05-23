package form

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGeoSearch(t *testing.T) {
	t.Run("valid query", func(t *testing.T) {
		form := &GeoSearch{Query: "query:\"fooBar baz\" before:2019-01-15 dist:25000 lat:33.45343166666667"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal("err should be nil")
		}

		log.Debugf("%+v\n", form)

		assert.Equal(t, "fooBar baz", form.Query)
		assert.Equal(t, time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), form.Before)
		assert.Equal(t, uint(0x61a8), form.Dist)
		assert.Equal(t, float32(33.45343), form.Lat)
	})
}

func TestNewGeoSearch(t *testing.T) {
	r := NewGeoSearch("Berlin")
	assert.IsType(t, GeoSearch{}, r)
}
