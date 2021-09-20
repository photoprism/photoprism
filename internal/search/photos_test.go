package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotos(t *testing.T) {
	t.Run("Chinese", func(t *testing.T) {
		var frm form.PhotoSearch

		frm.Query = "张"
		frm.Count = 10
		frm.Offset = 0

		_, _, err := Photos(frm)

		assert.NoError(t, err)
	})
	t.Run("search all", func(t *testing.T) {
		var frm form.PhotoSearch

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
		var frm form.PhotoSearch

		frm.Query = ""
		frm.Count = 5000
		frm.Offset = 0
		frm.ID = "pt9jtdre2lvl0yh7"
		frm.Merged = true

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for ID with merged false", func(t *testing.T) {
		var frm form.PhotoSearch

		frm.Query = ""
		frm.Count = 5000
		frm.Offset = 0
		frm.ID = "pt9jtdre2lvl0yh7"
		frm.Merged = false

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("label query dog", func(t *testing.T) {
		var frm form.PhotoSearch

		frm.Query = "label:dog"
		frm.Count = 10
		frm.Offset = 0

		photos, _, err := Photos(frm)

		assert.Equal(t, "dog not found", err.Error())
		assert.Empty(t, photos)
	})
	t.Run("label query landscape", func(t *testing.T) {
		var frm form.PhotoSearch

		frm.Query = "label:landscape Order:relevance"
		frm.Count = 10
		frm.Offset = 0

		photos, _, err := Photos(frm)
		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(photos))
	})
	t.Run("invalid label query", func(t *testing.T) {
		var frm form.PhotoSearch

		frm.Query = "label:xxx"
		frm.Count = 10
		frm.Offset = 0

		photos, _, err := Photos(frm)

		assert.Error(t, err)
		assert.Empty(t, photos)

		if err != nil {
			assert.Equal(t, err.Error(), "xxx not found")
		}
	})
	t.Run("form.location true", func(t *testing.T) {
		var frm form.PhotoSearch

		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0
		frm.Geo = true

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(photos))
	})
	t.Run("form.location true and keyword", func(t *testing.T) {
		var frm form.PhotoSearch

		frm.Query = "bridge"
		frm.Count = 10
		frm.Offset = 0
		frm.Geo = true
		frm.Error = false

		photos, _, err := Photos(frm)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for keyword", func(t *testing.T) {
		var frm form.PhotoSearch

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
		var frm form.PhotoSearch

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
		var f form.PhotoSearch

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
		var f form.PhotoSearch

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
		var f form.PhotoSearch

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
		var f form.PhotoSearch

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
		var f form.PhotoSearch

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
		var f form.PhotoSearch

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
	t.Run("form.camera", func(t *testing.T) {
		var f form.PhotoSearch

		f.Query = ""
		f.Count = 10
		f.Offset = 0
		f.Camera = 1000003

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 3, len(photos))
	})
	t.Run("form.color", func(t *testing.T) {
		var f form.PhotoSearch
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
		var f form.PhotoSearch
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
		var f form.PhotoSearch
		f.Query = "country:zz"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))

	})
	t.Run("form.title", func(t *testing.T) {
		var f form.PhotoSearch
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
		var f form.PhotoSearch
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
		var f form.PhotoSearch
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
		var f form.PhotoSearch
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
		var f form.PhotoSearch
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
		var f form.PhotoSearch
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
		var f form.PhotoSearch
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
		var f form.PhotoSearch
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
		var f form.PhotoSearch
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
	t.Run("form.chroma >9 Order:similar", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "chroma:25 Order:similar"
		f.Count = 3
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(photos))

	})
	t.Run("form.chroma <9", func(t *testing.T) {
		var f form.PhotoSearch
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
	t.Run("form.fmin and Order:oldest", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "Fmin:5 Order:oldest"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("form.fmax and Order:newest", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "Fmax:2 Order:newest"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(photos))

	})
	t.Run("form.Lat and form.Lng and Order:imported", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "Lat:33.45343166666667 Lng:25.764711666666667 Dist:2000 Order:imported"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 2, len(photos))

	})
	t.Run("form.Lat and form.Lng and Order:imported Dist:6000", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "Lat:33.45343166666667 Lng:25.764711666666667 Dist:6000 Order:imported"
		f.Count = 10
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 2, len(photos))

	})
	t.Run("form.Before and form.After Order:relevance", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "Before:2016-01-01 After:2013-01-01 Order:relevance"
		f.Count = 5000
		f.Offset = 0
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 2, len(photos))
	})

	t.Run("search for diff", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "Diff:800"
		f.Count = 5000
		f.Offset = 0

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for lens, month, year, album", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = ""
		f.Count = 5000
		f.Offset = 0
		f.Lens = 1000000
		f.Month = 7
		f.Year = 2790
		f.Album = "at9lxuqxpogaaba8"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("albums", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = ""
		f.Albums = "Berlin"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for state", func(t *testing.T) {
		var f form.PhotoSearch
		f.State = "KwaZulu-Natal"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})
	t.Run("search for category", func(t *testing.T) {
		var f form.PhotoSearch
		f.Category = "botanical garden"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})

	t.Run("search for labels", func(t *testing.T) {
		var f form.PhotoSearch
		f.Label = "botanical-garden|nature|landscape|park"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})

	t.Run("search for primary files", func(t *testing.T) {
		var f form.PhotoSearch
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})

	t.Run("search for landscape", func(t *testing.T) {
		var f form.PhotoSearch
		f.Query = "landscape"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(photos))
	})

	t.Run("search with multiple parameters", func(t *testing.T) {
		var f form.PhotoSearch
		f.Hidden = true
		f.Scan = true
		f.Year = 2010
		f.Day = 1
		f.Photo = true
		f.Name = "xxx"
		f.Original = "xxyy"
		f.Path = "/xxx/xxx/"
		f.Order = entity.SortOrderName

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, PhotoResults{}, photos)
	})
	t.Run("search with multiple parameters", func(t *testing.T) {
		var f form.PhotoSearch
		f.Hidden = true
		f.Scan = true
		f.Year = 2010
		f.Day = 1
		f.Video = true
		f.Name = "xxx"
		f.Original = "xxyy"
		f.Path = "/xxx|xxx"
		f.Type = "mp4"
		f.Stackable = true
		f.Unsorted = true
		f.Filter = ""
		f.Order = entity.SortOrderAdded

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.IsType(t, PhotoResults{}, photos)
	})
	t.Run("search all recently edited", func(t *testing.T) {
		var frm form.PhotoSearch

		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0
		frm.Order = entity.SortOrderEdited

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
		var frm form.PhotoSearch

		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0
		frm.Panorama = true
		frm.Stackable = false
		frm.Unstacked = true

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
		var frm form.PhotoSearch

		frm.Query = ""
		frm.Count = 10
		frm.Offset = 0
		frm.Name = "xxx|PhotoWithEditedAt"
		frm.Filename = "xxx|2007/12/PhotoWithEditedAt.jpg"
		frm.Title = "xxx|photowitheditedatdate"
		frm.Hash = "xxx|pcad9a68fa6acc5c5ba965adf6ec465ca42fd887"

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
		var f form.PhotoSearch
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
		var f form.PhotoSearch
		f.Faces = "Yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 3)
	})
	t.Run("faces:no", func(t *testing.T) {
		var f form.PhotoSearch
		f.Faces = "No"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 9)
	})
	t.Run("faces:2", func(t *testing.T) {
		var f form.PhotoSearch
		f.Faces = "2"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)
	})
	t.Run("filename", func(t *testing.T) {
		var f form.PhotoSearch
		f.Filename = "1990/04/Quality1FavoriteTrue.jpg"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 1)
	})
	t.Run("original name or original name", func(t *testing.T) {
		var f form.PhotoSearch
		f.Original = "my-videos/IMG_88888" + "|" + "Vacation/exampleFileNameOriginal"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 2)
	})
	t.Run("Stack", func(t *testing.T) {
		var f form.PhotoSearch
		f.Stack = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(photos), 2)
	})
}
