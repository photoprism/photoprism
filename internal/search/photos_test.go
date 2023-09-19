package search

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/sortby"
)

func TestPhotos(t *testing.T) {
	t.Run("OrderDuration", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0
		frm.Order = sortby.Duration

		photos, _, err := Photos(frm)
		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(photos))
	})
	t.Run("OrderRandom", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0
		frm.Order = sortby.Random

		photos, _, err := Photos(frm)
		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(photos))
	})
	t.Run("OrderInvalid", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0
		frm.Order = sortby.Invalid

		_, _, err := Photos(frm)
		assert.Error(t, err)
	})
	t.Run("Chinese", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "张"
		frm.Count = 10
		frm.Offset = 0

		_, _, err := Photos(frm)

		assert.NoError(t, err)
	})
	t.Run("UnknownFaces", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = ""
		frm.Face = "None"
		frm.Count = 10
		frm.Offset = 0

		results, _, err := Photos(frm)

		assert.NoError(t, err)
		assert.LessOrEqual(t, 1, len(results))
	})
	t.Run("search all", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 3, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.CameraID)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("search for ID and merged", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = ""
		frm.Count = 5000
		frm.Offset = 0
		frm.UID = "pt9jtdre2lvl0yh7"
		frm.Merged = true

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for ID with merged false", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = ""
		frm.Count = 5000
		frm.Offset = 0
		frm.UID = "pt9jtdre2lvl0yh7"
		frm.Merged = false

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("label query dog", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "label:dog"
		frm.Count = 10
		frm.Offset = 0

		photos, count, err := Photos(frm)

		assert.NoError(t, err)
		assert.Equal(t, PhotoResults{}, photos)
		assert.Equal(t, 0, count)
	})
	t.Run("label query landscape", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "label:landscape"
		frm.Count = 10
		frm.Offset = 0
		frm.Order = "relevance"

		photos, _, err := Photos(frm)
		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(photos))
	})
	t.Run("invalid label query", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "label:xxx"
		frm.Count = 10
		frm.Offset = 0

		photos, count, err := Photos(frm)

		assert.NoError(t, err)
		assert.Equal(t, PhotoResults{}, photos)
		assert.Equal(t, 0, count)
	})
	t.Run("form.location true", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0
		frm.Geo = "yes"

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(photos))
	})
	t.Run("form.location true and keyword", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "bridge"
		frm.Count = 10
		frm.Offset = 0
		frm.Geo = "yes"
		frm.Error = false

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for keyword", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "bridge"
		frm.Count = 5000
		frm.Offset = 0

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(photos))
	})
	t.Run("search for label in query", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "flower"
		frm.Count = 5000
		frm.Offset = 0

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for archived", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = ""
		f.Count = 5000
		f.Offset = 0
		f.Archived = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for private", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = ""
		f.Count = 5000
		f.Offset = 0
		f.Private = true
		f.Error = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for public", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = ""
		f.Count = 5000
		f.Offset = 0
		f.Public = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 3, len(photos))
	})
	t.Run("search for review", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = ""
		f.Count = 5000
		f.Offset = 0
		f.Review = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for quality", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = ""
		f.Count = 5000
		f.Offset = 0
		f.Quality = 3
		f.Private = false

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for file error", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = ""
		f.Count = 5000
		f.Offset = 0
		f.Error = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("form.camera- name", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = ""
		f.Count = 10
		f.Offset = 0
		f.Camera = "Canon EOS 6D"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 3, len(photos))
	})
	t.Run("camera:\"Canon EOS 6D\"", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "camera:\"Canon EOS 6D\""
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 3, len(photos))
	})
	t.Run("form.camera- id", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = ""
		f.Count = 10
		f.Offset = 0
		f.Camera = "1000003"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 3, len(photos))
	})
	t.Run("form.color", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = ""
		f.Count = 3
		f.Offset = 0
		f.Color = "blue"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("form.favorites", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "favorite:true"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("form.country", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "country:zz"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

	})
	t.Run("form.city", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "city:Mandeni"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

	})
	t.Run("form.title", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "title:Neckarbrücke"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("results: %+v", photos)
		assert.GreaterOrEqual(t, len(photos), 1)

	})
	t.Run("form.keywords", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "keywords:bridge"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("results: %+v", photos)
		assert.GreaterOrEqual(t, len(photos), 4)
	})
	t.Run("form.face", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "face:PN6QO5INYTUSAATOFL43LL2ABAV5ACZK"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("results: %+v", photos)
		assert.Equal(t, 1, len(photos))
	})
	t.Run("form.subject", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "subject:jqu0xs11qekk9jx8"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("results: %+v", photos)
		assert.Equal(t, 1, len(photos))
	})
	t.Run("form.subjects", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "subjects:John"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("results: %+v", photos)
		assert.Equal(t, 1, len(photos))
	})
	t.Run("form.people", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "people:John"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("results: %+v", photos)
		assert.Equal(t, 1, len(photos))
	})
	t.Run("form.hash", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "hash:2cad9168fa6acc5c5c2965ddf6ec465ca42fd818"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		//t.Logf("results: %+v", photos)
		assert.Equal(t, 1, len(photos))
	})

	t.Run("form.portrait", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "portrait:true"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

	})

	t.Run("form.mono", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "mono:true"
		f.Count = 10
		f.Offset = 0
		f.Archived = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("form.chroma:25", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "chroma:25"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(photos))

	})
	t.Run("form.chroma <9", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "chroma:4"
		f.Count = 3
		f.Offset = 0
		f.Error = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

	})
	t.Run("form.fmin", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "Fmin:5"
		f.Count = 10
		f.Offset = 0
		f.Order = "oldest"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("form.fmax", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "Fmax:2"
		f.Count = 10
		f.Offset = 0
		f.Order = "newest"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(photos))

	})
	t.Run("form.Lat and form.Lng", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "Lat:33.45343166666667 Lng:25.764711666666667 Dist:2000"
		f.Count = 10
		f.Offset = 0
		f.Order = "imported"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 2, len(photos))

	})
	t.Run("form.Lat and form.Lng and Dist:6000", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "Lat:33.45343166666667 Lng:25.764711666666667 Dist:6000"
		f.Count = 10
		f.Offset = 0
		f.Order = "imported"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 2, len(photos))

	})
	t.Run("LatLng:33.453431,-180.0,49.519234,180.0", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "LatLng:33.453431,-180.0,49.519234,180.0"
		f.Count = 10
		f.Offset = 0
		f.Order = "imported"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, p := range photos {
			assert.GreaterOrEqual(t, float32(49.519234), p.PhotoLat)
			assert.LessOrEqual(t, float32(33.45343166666667), p.PhotoLat)
		}

		assert.LessOrEqual(t, 2, len(photos))

	})
	t.Run("LatLng:0.00,-30.123.0,49.519234,9.1001234", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "LatLng:0.00,-30.123.0,49.519234,9.1001234"
		f.Count = 10
		f.Offset = 0
		f.Order = "imported"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		for _, p := range photos {
			assert.GreaterOrEqual(t, float32(49.519234), p.PhotoLat)
			assert.LessOrEqual(t, float32(0.00), p.PhotoLat)
			assert.GreaterOrEqual(t, float32(9.1001234), p.PhotoLng)
			assert.LessOrEqual(t, float32(-30.123), p.PhotoLng)
		}

		assert.LessOrEqual(t, 10, len(photos))

	})
	t.Run("form.Before and form.After Order:relevance", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "Before:2016-01-01 After:2013-01-01"
		f.Count = 5000
		f.Offset = 0
		f.Merged = true
		f.Order = "relevance"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 2, len(photos))
	})

	t.Run("search for diff", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "Diff:800"
		f.Count = 5000
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for camera name", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = ""
		f.Count = 1
		f.Offset = 0
		f.Camera = "canon"
		f.Lens = ""

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for lens name", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = ""
		f.Count = 1
		f.Offset = 0
		f.Camera = ""
		f.Lens = "apple"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for full lens name", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = ""
		f.Count = 1
		f.Offset = 0
		f.Camera = ""
		f.Lens = "Apple F380"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for full lens name using query", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "lens:\"Apple F380\""
		f.Count = 1
		f.Offset = 0
		f.Camera = ""

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for lens, month, year, album", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = ""
		f.Count = 5000
		f.Offset = 0
		f.Lens = "1000000"
		f.Month = strconv.Itoa(7)
		f.Year = strconv.Itoa(2790)
		f.Album = "at9lxuqxpogaaba8"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("albums", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = ""
		f.Albums = "Berlin"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("f.album", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = ""
		f.Album = "Berlin"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for state", func(t *testing.T) {
		var f form.SearchPhotos
		f.State = "KwaZulu-Natal"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for city", func(t *testing.T) {
		var f form.SearchPhotos
		f.City = "Mandeni"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for category", func(t *testing.T) {
		var f form.SearchPhotos
		f.Category = "botanical garden"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})

	t.Run("search for labels", func(t *testing.T) {
		var f form.SearchPhotos
		f.Label = "botanical-garden|nature|landscape|park"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})

	t.Run("search for primary files", func(t *testing.T) {
		var f form.SearchPhotos
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})

	t.Run("search for landscape", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "landscape"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})

	t.Run("search with multiple parameters", func(t *testing.T) {
		var f form.SearchPhotos
		f.Hidden = true
		f.Scan = "true"
		f.Year = "2010"
		f.Day = "1"
		f.Photo = true
		f.Name = "xxx"
		f.Original = "xxyy"
		f.Path = "/xxx/xxx/"
		f.Order = sortby.Name

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, PhotoResults{}, photos)
	})
	t.Run("search with multiple parameters", func(t *testing.T) {
		var f form.SearchPhotos
		f.Hidden = true
		f.Scan = "true"
		f.Year = strconv.Itoa(2010)
		f.Day = strconv.Itoa(1)
		f.Video = true
		f.Name = "xxx"
		f.Original = "xxyy"
		f.Path = "/xxx|xxx"
		f.Type = "mp4"
		f.Stackable = true
		f.Unsorted = true
		f.Filter = ""
		f.Order = sortby.Added

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, PhotoResults{}, photos)
	})
	t.Run("search all recently edited", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0
		frm.Order = sortby.Edited

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.CameraID)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("search unstacked panoramas", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0
		frm.Panorama = true
		frm.Stackable = false
		frm.Unstacked = true

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.CameraID)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("Or search", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0
		frm.Name = "xxx|PhotoWithEditedAt"
		frm.Filename = "xxx|2007/12/PhotoWithEditedAt.jpg"
		frm.Title = "xxx|photowitheditedatdate"
		frm.Hash = "xxx|pcad9a68fa6acc5c5ba965adf6ec465ca42fd887"

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.CameraID)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("faces", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "faces:true"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 3)
	})
	t.Run("faces:yes", func(t *testing.T) {
		var f form.SearchPhotos
		f.Faces = "Yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 3)
	})
	t.Run("faces:no", func(t *testing.T) {
		var f form.SearchPhotos
		f.Faces = "No"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 9)
	})
	t.Run("f.face yes", func(t *testing.T) {
		var f form.SearchPhotos
		f.Face = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 9)
	})
	t.Run("faces:2", func(t *testing.T) {
		var f form.SearchPhotos
		f.Faces = "2"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)
	})
	t.Run("Subject", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Subject = "jqu0xs11qekk9jx8"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.CameraID)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("NewFaces", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Face = "new"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("query: videos", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "videos"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)

			if r.PhotoType != "video" {
				t.Error("type should be video only")
			}

			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("query: video", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "video"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)

			if r.PhotoType != "video" {
				t.Error("type should be video only")
			}

			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("query: live", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "live"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.Equal(t, "live", r.PhotoType)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("f.live", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Live = true
		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.Equal(t, "live", r.PhotoType)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("query: raws", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "raws"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.Equal(t, "raw", r.PhotoType)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("f.Raw", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Raw = true
		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.Equal(t, "raw", r.PhotoType)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("query: faces", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "faces"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.LessOrEqual(t, 1, r.PhotoFaces)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("query: faces", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "faces:new"
		frm.Face = ""
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.LessOrEqual(t, 1, r.PhotoFaces)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("query: people", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "people"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.LessOrEqual(t, 1, r.PhotoFaces)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("query: favorites", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "favorites"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.True(t, r.PhotoFavorite)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("query: stacks", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "stacks"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("query: panoramas", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "panoramas"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.Equal(t, true, r.PhotoPanorama)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("query: scans", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "scans"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.Equal(t, true, r.PhotoScan)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("query: monochrome", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "monochrome"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("query: mono", func(t *testing.T) {
		var frm form.SearchPhotos

		frm.Query = "mono"
		frm.Count = 10
		frm.Offset = 0

		// Parse query string and filter.
		if err := frm.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

		for _, r := range photos {
			assert.IsType(t, Photo{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.LensID)

			if fix, ok := entity.PhotoFixtures[r.PhotoName]; ok {
				assert.Equal(t, fix.PhotoName, r.PhotoName)
			}
		}
	})
	t.Run("filename", func(t *testing.T) {
		var f form.SearchPhotos
		f.Filename = "1990/04/Quality1FavoriteTrue.jpg"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)
	})
	t.Run("original name or original name", func(t *testing.T) {
		var f form.SearchPhotos
		f.Original = "my-videos/IMG_88888" + "|" + "Vacation/exampleFileNameOriginal"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 2)
	})
	t.Run("Stack", func(t *testing.T) {
		var f form.SearchPhotos
		f.Stack = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 2)
	})
	t.Run("keywords:kuh|bridge > keywords:bridge&kuh", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "keywords:kuh|bridge"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		f.Query = "keywords:bridge&kuh"

		photos2, _, err2 := Photos(f)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Greater(t, len(photos), len(photos2))
	})
	t.Run("AlbumsOrSearch", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "albums:Holiday|Berlin"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 5)
	})
	t.Run("AlbumsAndSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Berlin&Holiday\""

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 0)
	})
	t.Run("SubjectsAndOrSearch", func(t *testing.T) {
		var f form.SearchPhotos
		f.Subjects = "Actor A|Actress A"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		f.Subjects = "Actor A&Actress A"

		photos2, _, err2 := Photos(f)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Greater(t, len(photos), len(photos2))
	})
	t.Run("people = subjects & person = subject", func(t *testing.T) {
		var f form.SearchPhotos
		f.People = "Actor"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var f2 form.SearchPhotos

		f2.Subjects = "Actor"

		// Parse query string and filter.
		if err := f2.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos2, _, err2 := Photos(f2)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(photos2))

		var f3 form.SearchPhotos

		f3.Person = "Actor A"

		// Parse query string and filter.
		if err := f3.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos3, _, err3 := Photos(f3)

		if err3 != nil {
			t.Fatal(err3)
		}

		var f4 form.SearchPhotos
		f4.Subject = "Actor A"

		// Parse query string and filter.
		if err := f4.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		photos4, _, err4 := Photos(f4)

		if err4 != nil {
			t.Fatal(err4)
		}

		assert.Equal(t, len(photos3), len(photos4))
		assert.Equal(t, len(photos), len(photos4))
	})

	t.Run("Search in Title", func(t *testing.T) {
		var f form.SearchPhotos
		f.Query = "N"
		f.Title = ""
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", photos)
		assert.Equal(t, 1, len(photos))
		assert.Equal(t, photos[0].PhotoTitle, "Neckarbrücke")
	})
}
