package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func TestGeo(t *testing.T) {
	t.Run("Near", func(t *testing.T) {
		query := form.NewSearchPhotosGeo("near:ps6sg6be2lvl0y43")

		// Parse query string and filter.
		if err := query.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		if result, err := PhotosGeo(query); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("RESULT: %#v", result)
			assert.LessOrEqual(t, 4, len(result))
		}
	})
	t.Run("UnknownFaces", func(t *testing.T) {
		query := form.NewSearchPhotosGeo("face:none")

		// Parse query string and filter.
		if err := query.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		if result, err := PhotosGeo(query); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 0, len(result))
		}
	})
	t.Run("form.keywords", func(t *testing.T) {
		query := form.NewSearchPhotosGeo("keywords:bridge")

		// Parse query string and filter.
		if err := query.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		if result, err := PhotosGeo(query); err != nil {
			t.Fatal(err)
		} else {
			assert.GreaterOrEqual(t, len(result), 1)
		}
	})
	t.Run("form.subjects", func(t *testing.T) {
		query := form.NewSearchPhotosGeo("subjects:John")

		// Parse query string and filter.
		if err := query.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		if result, err := PhotosGeo(query); err != nil {
			t.Fatal(err)
		} else {
			assert.GreaterOrEqual(t, len(result), 0)
		}
	})
	t.Run("find_all", func(t *testing.T) {
		query := form.NewSearchPhotosGeo("")

		// Parse query string and filter.
		if err := query.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		if result, err := PhotosGeo(query); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 4, len(result))
		}
	})

	t.Run("search for bridge", func(t *testing.T) {
		query := form.NewSearchPhotosGeo("q:bridge Before:3006-01-02")

		// Parse query string and filter.
		if err := query.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		result, err := PhotosGeo(query)
		t.Logf("RESULT: %+v", result)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(result))

	})

	t.Run("search for date range", func(t *testing.T) {
		query := form.NewSearchPhotosGeo("After:2014-12-02 Before:3006-01-02")

		// Parse query string and filter.
		if err := query.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		result, err := PhotosGeo(query)

		// t.Logf("RESULT: %+v", result)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Reunion", result[0].PhotoTitle)
	})

	t.Run("search for review true, quality 0", func(t *testing.T) {
		f := form.SearchPhotosGeo{
			Query:    "",
			Before:   time.Time{},
			After:    time.Time{},
			Favorite: "true",
			Lat:      1.234,
			Lng:      4.321,
			S2:       "",
			Olc:      "",
			Dist:     0,
			Quality:  0,
			Review:   true,
		}

		result, err := PhotosGeo(f)

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
		f := form.SearchPhotosGeo{
			Query:    "",
			Before:   time.Time{},
			After:    time.Time{},
			Favorite: "false",
			Lat:      0,
			Lng:      0,
			S2:       "",
			Olc:      "",
			Dist:     0,
			Quality:  3,
			Review:   false,
		}

		result, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 3, len(result))
		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("search for min and max altitude", func(t *testing.T) {
		f := form.SearchPhotosGeo{
			Query:    "",
			Before:   time.Time{},
			After:    time.Time{},
			Favorite: "false",
			Lat:      0,
			Lng:      0,
			S2:       "",
			Olc:      "",
			Dist:     0,
			Alt:      "200-500",
		}

		result, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("search for s2", func(t *testing.T) {
		f := form.SearchPhotosGeo{
			Query:    "",
			Before:   time.Time{},
			After:    time.Time{},
			Favorite: "false",
			Lat:      0,
			Lng:      0,
			S2:       "85",
			Olc:      "",
			Dist:     0,
			Quality:  0,
			Review:   false,
		}

		result, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, result)
		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("search for OLC", func(t *testing.T) {
		f := form.SearchPhotosGeo{
			Query:    "",
			Before:   time.Time{},
			After:    time.Time{},
			Favorite: "false",
			Lat:      0,
			Lng:      0,
			S2:       "",
			Olc:      "9",
			Dist:     0,
			Quality:  0,
			Review:   false,
		}

		result, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("query for label flower", func(t *testing.T) {
		f := form.SearchPhotosGeo{
			Query: "flower",
		}

		result, err := PhotosGeo(f)
		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("query for label landscape", func(t *testing.T) {
		f := form.SearchPhotosGeo{
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

		result, err := PhotosGeo(f)
		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("search with multiple parameters", func(t *testing.T) {
		f := form.SearchPhotosGeo{
			Query:    "landscape",
			Photo:    true,
			Path:     "/xxx,xxx",
			Name:     "xxx",
			Archived: false,
			Private:  false,
			Public:   true,
		}

		result, err := PhotosGeo(f)
		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("search for archived true", func(t *testing.T) {
		f := form.SearchPhotosGeo{
			Query:    "landscape",
			Photo:    true,
			Path:     "/xxx/xxx/",
			Name:     "xxx",
			Archived: true,
		}

		result, err := PhotosGeo(f)
		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, GeoResults{}, result)
	})
	t.Run("faces:true", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Query = "faces:true"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 4)
	})
	t.Run("faces:yes", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Faces = "Yes"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 4)
	})
	t.Run("face:yes", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Face = "Yes"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 4)
	})
	t.Run("f.Faces:new", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Faces = "New"
		f.Face = ""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		// TODO: Should be 3 or more, check entity fixtures!
		assert.GreaterOrEqual(t, len(photos), 2)
	})
	t.Run("faces:no", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Faces = "No"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 8)
	})
	t.Run("faces:2", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Faces = "2"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)
	})
	t.Run("face: TOSCDXCS4VI3PGIUTCNIQCNI6HSFXQVZ", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Face = "TOSCDXCS4VI3PGIUTCNIQCNI6HSFXQVZ"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 2)
	})
	t.Run("day", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Day = "18"
		f.Month = "4"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)
	})
	t.Run("subject uid in query", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Query = "Actress"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)
	})
	t.Run("Album", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Album = "Berlin"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(photos))
	})
	t.Run("Albums", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Albums = "Holiday|Christmas"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(photos))
	})
	t.Run("City", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.City = "Teotihuacán"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 8, len(photos))
	})
	t.Run("PathOrPath", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Path = "1990/04" + "|" + "2015/11"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 3)
	})
	t.Run("name or name", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Name = "20151101_000000_51C501B5" + "|" + "Video"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 2)
	})
	t.Run("query: videos", func(t *testing.T) {
		var frm form.SearchPhotosGeo

		frm.Query = "videos"

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo

		frm.Query = "video"

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo

		frm.Query = "live"

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo

		frm.Query = "raws"

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo

		frm.Query = "panoramas"

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo

		frm.Query = "scans"

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo

		frm.Query = "faces"

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo

		frm.Query = "people"

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo

		frm.Query = "favorites"

		photos, err := PhotosGeo(frm)

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
		var f form.SearchPhotosGeo
		f.Query = "keywords:kuh|bridge"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		f.Query = "keywords:bridge&kuh"

		photos2, err2 := PhotosGeo(f)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Greater(t, len(photos), len(photos2))
	})
	t.Run("albums and and or search", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Query = "albums:Holiday|Berlin"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		f.Query = "albums:\"Berlin&Holiday|Christmas\""

		photos2, err2 := PhotosGeo(f)

		if err2 != nil {
			t.Fatal(err2)
		}
		assert.Greater(t, len(photos), len(photos2))
	})
	t.Run("f.Album = uid", func(t *testing.T) {
		var frm form.SearchPhotosGeo

		frm.Album = "as6sg6bxpogaaba9"

		photos, err := PhotosGeo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
	t.Run("subjects and and or search", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.Subjects = "Actor A|Actress A"

		t.Logf("S1: %s", f.SerializeAll())

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		f.Subjects = "Actor A&Actress A"

		t.Logf("S2: %s", f.SerializeAll())

		photos2, err2 := PhotosGeo(f)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Greater(t, len(photos), len(photos2))
	})
	t.Run("people = subjects & person = subject", func(t *testing.T) {
		var f form.SearchPhotosGeo
		f.People = "Actor"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		var f2 form.SearchPhotosGeo

		f2.Subjects = "Actor"

		// Parse query string and filter.
		if err = f2.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos2, err2 := PhotosGeo(f2)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(photos2))

		var f3 form.SearchPhotosGeo

		f3.Person = "Actor A"

		// Parse query string and filter.
		if err = f3.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos3, err3 := PhotosGeo(f3)

		if err3 != nil {
			t.Fatal(err3)
		}

		var f4 form.SearchPhotosGeo
		f4.Subject = "Actor A"

		// Parse query string and filter.
		if err = f4.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos4, err4 := PhotosGeo(f4)

		if err4 != nil {
			t.Fatal(err4)
		}

		assert.Equal(t, len(photos3), len(photos4))
		assert.Equal(t, len(photos), len(photos4))

		var f5 form.SearchPhotosGeo
		f5.Subject = "js6sg6b1h1njaaad"

		// Parse query string and filter.
		if err = f5.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos5, err5 := PhotosGeo(f5)

		if err5 != nil {
			t.Fatal(err5)
		}

		assert.Equal(t, len(photos5), len(photos4))
	})
	t.Run("f.Scan = true", func(t *testing.T) {
		var frm form.SearchPhotosGeo

		frm.Scan = "true"

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo

		frm.Panorama = true

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo

		frm.Raw = true

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo

		frm.Live = true

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo

		frm.Title = "phototobebatchapproved2"

		photos, err := PhotosGeo(frm)

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
		var frm form.SearchPhotosGeo
		frm.Query = "p"
		frm.Title = ""

		photos, err := PhotosGeo(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, GeoResult{}, r)
			assert.NotEmpty(t, r.ID)
		}
	})
	t.Run("Panorama", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "panorama:true"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("Portrait", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "portrait:true"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("Landscape", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "landscape:true"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("Square", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "square:true"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
}
