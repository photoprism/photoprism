package query

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestGeo(t *testing.T) {
	t.Run("search all photos", func(t *testing.T) {
		query := form.NewGeoSearch("")
		result, err := Geo(query)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 5, len(result))
	})

	t.Run("search for bridge", func(t *testing.T) {
		query := form.NewGeoSearch("Query:bridge Before:3006-01-02")
		result, err := Geo(query)
		t.Logf("RESULT: %+v", result)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(result))

	})

	t.Run("search for date range", func(t *testing.T) {
		query := form.NewGeoSearch("After:2014-12-02 Before:3006-01-02")
		result, err := Geo(query)

		// t.Logf("RESULT: %+v", result)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Reunion", result[0].PhotoTitle)
	})

	t.Run("search for review true, quality 0", func(t *testing.T) {
		f := form.GeoSearch{
			Query:    "",
			Before:   time.Time{},
			After:    time.Time{},
			Favorite: true,
			Lat:      1.234,
			Lng:      4.321,
			S2:       "",
			Olc:      "",
			Dist:     0,
			Quality:  0,
			Review:   true,
		}

		result, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(result))
		assert.Equal(t, "1000017", result[0].ID)
		assert.IsType(t, GeoResults{}, result)
	})

	t.Run("search for review false, quality > 0", func(t *testing.T) {
		f := form.GeoSearch{
			Query:    "",
			Before:   time.Time{},
			After:    time.Time{},
			Favorite: false,
			Lat:      0,
			Lng:      0,
			S2:       "",
			Olc:      "",
			Dist:     0,
			Quality:  3,
			Review:   false,
		}

		result, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 4, len(result))
		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("search for s2", func(t *testing.T) {
		f := form.GeoSearch{
			Query:    "",
			Before:   time.Time{},
			After:    time.Time{},
			Favorite: false,
			Lat:      0,
			Lng:      0,
			S2:       "85",
			Olc:      "",
			Dist:     0,
			Quality:  0,
			Review:   false,
		}

		result, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, result)
		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("search for Olc", func(t *testing.T) {
		f := form.GeoSearch{
			Query:    "",
			Before:   time.Time{},
			After:    time.Time{},
			Favorite: false,
			Lat:      0,
			Lng:      0,
			S2:       "",
			Olc:      "9",
			Dist:     0,
			Quality:  0,
			Review:   false,
		}

		result, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("query too short", func(t *testing.T) {
		f := form.GeoSearch{
			Query: "a",
		}

		result, err := Geo(f)

		assert.Error(t, err)
		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("query for label flower", func(t *testing.T) {
		f := form.GeoSearch{
			Query: "flower",
		}

		result, err := Geo(f)
		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("query for label landscape", func(t *testing.T) {
		f := form.GeoSearch{
			Query:    "landscape",
			Album:    "test",
			Camera:   123,
			Lens:     123,
			Year:     2010,
			Month:    12,
			Color:    "red",
			Country:  "zz",
			Type:     "jpg",
			Video:    true,
			Path:     "/xxx/xxx/",
			Name:     "xxx",
			Archived: false,
			Private:  true,
		}

		result, err := Geo(f)
		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("search with multiple parameters", func(t *testing.T) {
		f := form.GeoSearch{
			Query:    "landscape",
			Photo:    true,
			Path:     "/xxx,xxx",
			Name:     "xxx",
			Archived: false,
			Private:  false,
			Public:   true,
		}

		result, err := Geo(f)
		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("search for archived true", func(t *testing.T) {
		f := form.GeoSearch{
			Query:    "landscape",
			Photo:    true,
			Path:     "/xxx/xxx/",
			Name:     "xxx",
			Archived: true,
		}

		result, err := Geo(f)
		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, GeoResults{}, result)
	})
}
