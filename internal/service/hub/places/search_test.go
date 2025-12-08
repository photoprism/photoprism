package places

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	lat, lng := 52.5108869, 13.398947

	t.Run("Success", func(t *testing.T) {
		l, err := Search("Berlin", "de", 10)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, l, 1)

		if len(l) > 0 {
			assert.Equal(t, "Berlin", l[0].City)
			assert.Equal(t, "de", l[0].Country)
			assert.Equal(t, lat, l[0].Lat)
			assert.Equal(t, lng, l[0].Lng)
		}
	})
	t.Run("MissingQuery", func(t *testing.T) {
		l, err := Search("", "", 10)
		assert.Len(t, l, 0)
		assert.Error(t, err, "places: invalid location id ")
	})
	t.Run("Cached", func(t *testing.T) {
		l, err := Search("Berlin, Deutschland", "de", 10)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, l, 1)

		if len(l) > 0 {
			assert.Equal(t, "Berlin", l[0].City)
			assert.Equal(t, "de", l[0].Country)
			assert.Equal(t, lat, l[0].Lat)
			assert.Equal(t, lng, l[0].Lng)
		}

		l, err = Search("Berlin, Deutschland", "de", 10)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, l, 1)

		if len(l) > 0 {
			assert.Equal(t, "Berlin", l[0].City)
			assert.Equal(t, "de", l[0].Country)
		}
	})
}
