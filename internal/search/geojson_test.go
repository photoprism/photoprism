package search

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestGeo(t *testing.T) {
	t.Run("Near", func(t *testing.T) {
		query := form.NewGeoSearch("near:pt9jtdre2lvl0y43")

		if result, err := Geo(query); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("RESULT: %#v", result)
			assert.LessOrEqual(t, 4, len(result))
		}
	})
	t.Run("UnknownFaces", func(t *testing.T) {
		query := form.NewGeoSearch("face:none")

		if result, err := Geo(query); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 0, len(result))
		}
	})
	t.Run("form.keywords", func(t *testing.T) {
		query := form.NewGeoSearch("keywords:bridge")

		if result, err := Geo(query); err != nil {
			t.Fatal(err)
		} else {
			assert.GreaterOrEqual(t, len(result), 1)
		}
	})
	t.Run("form.subjects", func(t *testing.T) {
		query := form.NewGeoSearch("subjects:John")

		if result, err := Geo(query); err != nil {
			t.Fatal(err)
		} else {
			assert.GreaterOrEqual(t, len(result), 0)
		}
	})
	t.Run("find_all", func(t *testing.T) {
		query := form.NewGeoSearch("")

		if result, err := Geo(query); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 4, len(result))
		}
	})

	t.Run("search for bridge", func(t *testing.T) {
		query := form.NewGeoSearch("Query:bridge Before:3006-01-02")
		result, err := Geo(query)
		t.Logf("RESULT: %+v", result)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(result))

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
		f := form.SearchGeo{
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
		assert.IsType(t, GeoResults{}, result)

		if len(result) > 0 {
			assert.Equal(t, "1000017", result[0].ID)
		}
	})

	t.Run("search for review false, quality > 0", func(t *testing.T) {
		f := form.SearchGeo{
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
		assert.LessOrEqual(t, 3, len(result))
		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("search for s2", func(t *testing.T) {
		f := form.SearchGeo{
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
		f := form.SearchGeo{
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
	t.Run("query for label flower", func(t *testing.T) {
		f := form.SearchGeo{
			Query: "flower",
		}

		result, err := Geo(f)
		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("query for label landscape", func(t *testing.T) {
		f := form.SearchGeo{
			Query:    "landscape",
			Album:    "test",
			Camera:   123,
			Lens:     123,
			Year:     "2010",
			Month:    "12",
			Color:    "red",
			Country:  entity.UnknownID,
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
		f := form.SearchGeo{
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
		f := form.SearchGeo{
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
	t.Run("faces:true", func(t *testing.T) {
		var f form.SearchGeo
		f.Query = "faces:true"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 4)
	})
	t.Run("faces:yes", func(t *testing.T) {
		var f form.SearchGeo
		f.Faces = "Yes"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 4)
	})
	t.Run("face:yes", func(t *testing.T) {
		var f form.SearchGeo
		f.Face = "Yes"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 4)
	})
	t.Run("f.Faces:new", func(t *testing.T) {
		var f form.SearchGeo
		f.Faces = "New"
		f.Face = ""

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 3)
	})
	t.Run("faces:no", func(t *testing.T) {
		var f form.SearchGeo
		f.Faces = "No"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 8)
	})
	t.Run("faces:2", func(t *testing.T) {
		var f form.SearchGeo
		f.Faces = "2"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)
	})
	t.Run("face: TOSCDXCS4VI3PGIUTCNIQCNI6HSFXQVZ", func(t *testing.T) {
		var f form.SearchGeo
		f.Face = "TOSCDXCS4VI3PGIUTCNIQCNI6HSFXQVZ"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 2)
	})
	t.Run("day", func(t *testing.T) {
		var f form.SearchGeo
		f.Day = "18"
		f.Month = "4"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)
	})
	t.Run("subject uid in query", func(t *testing.T) {
		var f form.SearchGeo
		f.Query = "Actress"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)
	})
	t.Run("albums", func(t *testing.T) {
		var f form.SearchGeo
		f.Albums = "2030"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 10)
	})
	t.Run("path or path", func(t *testing.T) {
		var f form.SearchGeo
		f.Path = "1990/04" + "|" + "2015/11"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 3)
	})
	t.Run("name or name", func(t *testing.T) {
		var f form.SearchGeo
		f.Name = "20151101_000000_51C501B5" + "|" + "Video"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 2)
	})
	t.Run("query: videos", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Query = "videos"

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
			assert.Equal(t, "video", r.PhotoType)
		}
	})
	t.Run("query: video", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Query = "video"

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
			assert.Equal(t, "video", r.PhotoType)
		}
	})
	t.Run("query: live", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Query = "live"

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
			assert.Equal(t, "live", r.PhotoType)
		}
	})
	t.Run("query: raws", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Query = "raws"

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
			assert.Equal(t, "raw", r.PhotoType)
		}
	})
	t.Run("query: panoramas", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Query = "panoramas"

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
	t.Run("query: scans", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Query = "scans"

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
	t.Run("query: faces", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Query = "faces"

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
	t.Run("query: people", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Query = "people"

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
	t.Run("query: favorites", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Query = "favorites"

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
			assert.True(t, r.PhotoFavorite)
		}
	})
	t.Run("keywords:kuh|bridge > keywords:bridge&kuh", func(t *testing.T) {
		var f form.SearchGeo
		f.Query = "keywords:kuh|bridge"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		f.Query = "keywords:bridge&kuh"

		photos2, err2 := Geo(f)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Greater(t, len(photos), len(photos2))
	})
	t.Run("albums and and or search", func(t *testing.T) {
		var f form.SearchGeo
		f.Query = "albums:Holiday|Berlin"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		f.Query = "albums:Berlin&Holiday"

		photos2, err2 := Geo(f)

		if err2 != nil {
			t.Fatal(err2)
		}
		assert.Greater(t, len(photos), len(photos2))
	})
	t.Run("f.Album = uid", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Album = "at9lxuqxpogaaba9"

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
	t.Run("people and and or search", func(t *testing.T) {
		var f form.SearchGeo
		f.People = "Actor A|Actress A"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}

		f.People = "Actor A&Actress A"

		photos2, err2 := Geo(f)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Greater(t, len(photos), len(photos2))
	})
	t.Run("people = subjects & person = subject", func(t *testing.T) {
		var f form.SearchGeo
		f.People = "Actor"

		photos, err := Geo(f)

		if err != nil {
			t.Fatal(err)
		}
		var f2 form.SearchGeo

		f2.Subjects = "Actor"

		photos2, err2 := Geo(f2)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(photos2))

		var f3 form.SearchGeo

		f3.Person = "Actor A"

		photos3, err3 := Geo(f3)

		if err3 != nil {
			t.Fatal(err3)
		}

		var f4 form.SearchGeo
		f4.Subject = "Actor A"

		photos4, err4 := Geo(f4)

		if err4 != nil {
			t.Fatal(err4)
		}

		assert.Equal(t, len(photos3), len(photos4))
		assert.Equal(t, len(photos), len(photos4))

		var f5 form.SearchGeo
		f5.Subject = "jqy1y111h1njaaad"

		photos5, err5 := Geo(f5)

		if err5 != nil {
			t.Fatal(err5)
		}

		assert.Equal(t, len(photos5), len(photos4))
	})

	t.Run("f.Scan = true", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Scan = true

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
	t.Run("f.Panorama = true", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Panorama = true

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
	t.Run("f.Raw = true", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Raw = true

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
	t.Run("f.Live = true", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Live = true

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
	t.Run("f.Title = phototobebatchapproved2", func(t *testing.T) {
		var frm form.SearchGeo

		frm.Title = "phototobebatchapproved2"

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
	t.Run("QueryP", func(t *testing.T) {
		var frm form.SearchGeo
		frm.Query = "p"
		frm.Title = ""

		photos, err := Geo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
}
