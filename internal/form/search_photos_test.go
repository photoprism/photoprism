package form

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/pkg/txt"
)

func TestSearchPhotosForm(t *testing.T) {
	form := &SearchPhotos{}

	assert.IsType(t, new(SearchPhotos), form)
}

func TestParseQueryString(t *testing.T) {
	t.Run("Subjects", func(t *testing.T) {
		form := &SearchPhotos{Query: "subjects:\"Jens & Mander\""}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Jens & Mander", form.Subjects)
	})
	t.Run("Subject", func(t *testing.T) {
		form := &SearchPhotos{Query: "subject:\"Jens\""}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Jens", form.Subject)
		assert.Equal(t, "", form.Person)
	})
	t.Run("Aliases", func(t *testing.T) {
		form := &SearchPhotos{Query: "people:\"Jens & Mander\" folder:Foo person:Bar"}

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
	t.Run("Keywords", func(t *testing.T) {
		form := &SearchPhotos{Query: "keywords:\"Foo Bar\""}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Foo Bar", form.Keywords)
	})
	t.Run("AndQuery", func(t *testing.T) {
		form := &SearchPhotos{Query: "\"Jens & Mander\" title:\"Tübingen\""}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "jens & mander", form.GetQuery())
		assert.Equal(t, "Tübingen", form.Title)
	})
	t.Run("Path", func(t *testing.T) {
		form := &SearchPhotos{Query: "path:123abc/,EFG"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "123abc/,EFG", form.Path)
	})
	t.Run("Folder", func(t *testing.T) {
		form := &SearchPhotos{Query: "folder:123abc/,EFG"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "123abc/,EFG", form.Path)
	})
	t.Run("ValidQuery", func(t *testing.T) {
		form := &SearchPhotos{Query: "label:cat q:\"fooBar baz\" before:2019-01-15 camera:23 favorite:false dist:25000 lat:33.45343166666667"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "cat", form.Label)
		assert.Equal(t, "fooBar baz", form.Query)
		assert.Equal(t, "23", form.Camera)
		assert.Equal(t, time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), form.Before)
		assert.Equal(t, "false", form.Favorite)
		assert.Equal(t, float64(25000), form.Dist)
		assert.Equal(t, 33.45343166666667, form.Lat)
	})
	t.Run("ValidQueryTwo", func(t *testing.T) {
		form := &SearchPhotos{Query: "chroma:200 title:\"te:st\" after:2018-01-15 favorite:true lng:33.45343166666667"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal("err should be nil")
		}

		assert.Equal(t, int16(200), form.Chroma)
		assert.Equal(t, "te:st", form.Title)
		assert.Equal(t, time.Date(2018, 01, 15, 0, 0, 0, 0, time.UTC), form.After)
		assert.Equal(t, 33.45343166666667, form.Lng)
	})
	t.Run("ValidQueryWithFilter", func(t *testing.T) {
		form := &SearchPhotos{Query: "label:cat title:\"fooBar baz\"", Filter: "label:dog"}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "dog", form.Label)
		assert.Equal(t, "label:dog", form.Filter)
		assert.Equal(t, "fooBar baz", form.Title)
	})
	t.Run("ValidQueryWithUmlauts", func(t *testing.T) {
		form := &SearchPhotos{Query: "title:\"tübingen\""}

		err := form.ParseQueryString()

		// log.Debugf("%+v\n", form)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "tübingen", form.Title)
	})
	t.Run("QueryForInvalidFilter", func(t *testing.T) {
		form := &SearchPhotos{Query: "xxx:false"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		// log.Debugf("%+v\n", form)

		assert.Equal(t, "unknown filter: xxx", err.Error())
	})
	t.Run("QueryForFavoritesWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "favorite:cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "cat", form.Favorite)
	})
	t.Run("QueryForPrimaryWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "primary:&cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Primary)
	})
	t.Run("QueryForStackWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "stack:*cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Stack)
	})
	t.Run("QueryForUnstackedWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "unstacked:'cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Unstacked)
	})
	t.Run("QueryForStackableWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "stackable:mother's day"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Stackable)
	})
	t.Run("QueryForVideoWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "video:|cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Video)
	})
	t.Run("AnimatedYes", func(t *testing.T) {
		form := &SearchPhotos{Query: "animated:yes"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, form.Vector)
		assert.True(t, form.Animated)
	})
	t.Run("VectorYes", func(t *testing.T) {
		form := &SearchPhotos{Query: "vector:yes"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, form.Animated)
		assert.True(t, form.Vector)
	})
	t.Run("QueryForPhotoWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "photo:cat>"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Photo)
	})
	t.Run("QueryForRawWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "raw:ca+(t"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Raw)
	})
	t.Run("QueryForLiveWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "live:cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Live)
	})
	t.Run("QueryForScanWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "scan:;cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, ";cat", form.Scan)
	})
	t.Run("QueryForPanoramaWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "panorama:*cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Panorama)
	})
	t.Run("QueryForErrorWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "error:^cat$#"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Error)
	})
	t.Run("QueryForHiddenWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "hidden:*cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Hidden)
	})
	t.Run("QueryForArchivedWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "archived:`cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Archived)
	})
	t.Run("QueryForPublicWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "public:*cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Public)
	})
	t.Run("QueryForPrivateWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "private:*c@t"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Private)
	})
	t.Run("QueryForUnsortedWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "unsorted:*cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Unsorted)
	})
	t.Run("QueryForMonoWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "mono:*cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Mono)
	})
	t.Run("QueryForPortraitWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "portrait:*cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Portrait)
	})
	t.Run("PortraitLandscapeSquare", func(t *testing.T) {
		form := &SearchPhotos{Query: "portrait:true landscape:yes square:jo"}

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
	t.Run("QueryForGeoWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "geo:*cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "*cat", form.Geo)
		assert.False(t, txt.No(form.Geo))
	})
	t.Run("QueryForReviewWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "review:*cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Review)
	})
	t.Run("QueryForMergedWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "merged:*cat"}

		err := form.ParseQueryString()

		assert.Error(t, err)
	})
	t.Run("QueryForLandscapeWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "landscape:test$5123"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Landscape)
	})
	t.Run("QueryForSquareWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "square:%abc"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Square)
	})
	t.Run("QueryForAnimatedWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "animated:%abc"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Animated)
	})
	t.Run("QueryForVectorWithUncommonBoolValue", func(t *testing.T) {
		form := &SearchPhotos{Query: "vector:%abc"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, form.Vector)
	})
	t.Run("QueryForLatWithInvalidType", func(t *testing.T) {
		form := &SearchPhotos{Query: "lat:&cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		// log.Debugf("%+v\n", form)

		assert.Contains(t, err.Error(), "invalid syntax")
	})
	t.Run("QueryForLngWithInvalidType", func(t *testing.T) {
		form := &SearchPhotos{Query: "lng:^>cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		// log.Debugf("%+v\n", form)

		assert.Contains(t, err.Error(), "invalid syntax")
	})
	t.Run("QueryForDistWithInvalidType", func(t *testing.T) {
		form := &SearchPhotos{Query: "dist:c@t"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		// log.Debugf("%+v\n", form)

		assert.Contains(t, err.Error(), "invalid syntax")
	})
	t.Run("QueryForFWithInvalidType", func(t *testing.T) {
		form := &SearchPhotos{Query: "f:=}cat{"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

	})
	t.Run("QueryForFWithInvalidType", func(t *testing.T) {
		form := &SearchPhotos{Query: "f:ca#$t"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("QueryForChromaWithInvalidType", func(t *testing.T) {
		form := &SearchPhotos{Query: "chroma:&|cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		// log.Debugf("%+v\n", form)

		assert.Contains(t, err.Error(), "invalid syntax")
	})
	t.Run("QueryForDiffWithInvalidType", func(t *testing.T) {
		form := &SearchPhotos{Query: "diff:&cat;%"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		// log.Debugf("%+v\n", form)

		assert.Contains(t, err.Error(), "invalid syntax")
	})
	t.Run("QueryForQualityWithInvalidType", func(t *testing.T) {
		form := &SearchPhotos{Query: "quality:`cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		// log.Debugf("%+v\n", form)

		assert.Contains(t, err.Error(), "invalid syntax")
	})
	t.Run("QueryForCountWithInvalidType", func(t *testing.T) {
		form := &SearchPhotos{Query: "dist:ca(%t"}

		err := form.ParseQueryString()

		assert.Error(t, err)
	})
	t.Run("QueryForOffsetWithInvalidType", func(t *testing.T) {
		form := &SearchPhotos{Query: "lat:&cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		// log.Debugf("%+v\n", form)

		assert.Contains(t, err.Error(), "invalid syntax")
	})
	t.Run("CameraString", func(t *testing.T) {
		form := &SearchPhotos{Query: "camera:cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "cat", form.Camera)
	})
	t.Run("LensString", func(t *testing.T) {
		form := &SearchPhotos{Query: "lens:cat"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "cat", form.Lens)
	})
	t.Run("Altitude", func(t *testing.T) {
		form := &SearchPhotos{Query: "alt:200-500"}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "200-500", form.Alt)
	})
	t.Run("BeforeTime", func(t *testing.T) {
		form := &SearchPhotos{Query: "before:time"}

		err := form.ParseQueryString()

		t.Logf("%s", form.Before)

		if err != nil {
			assert.Equal(t, "invalid before date", err.Error())
		} else {
			t.Fatal("'invalid before date' error expected")
		}
	})
	t.Run("BeforeCat", func(t *testing.T) {
		form := &SearchPhotos{Query: "before:cat"}

		err := form.ParseQueryString()

		if err != nil {
			assert.Equal(t, "invalid before date", err.Error())
		} else {
			t.Fatal("'invalid before date' error expected")
		}
	})
	t.Run("AfterCat", func(t *testing.T) {
		form := &SearchPhotos{Query: "after:cat"}

		err := form.ParseQueryString()

		if err == nil {
			t.Fatal(err)
		}

		// log.Debugf("%+v\n", form)

		assert.Equal(t, "invalid after date", err.Error())
	})
	t.Run("Id", func(t *testing.T) {
		form := &SearchPhotos{Query: "id:\"ii3e4567-e89b-hdgtr\""}

		err := form.ParseQueryString()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "ii3e4567-e89b-hdgtr", form.ID)
	})
	t.Run("Added", func(t *testing.T) {
		form := &SearchPhotos{Query: "added:\"2022-01-02T13:04:05+01:00\""}

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
		form := &SearchPhotos{Query: "updated:\"2001-01-02 17:04:05\""}

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
		form := &SearchPhotos{}

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

		assert.Equal(t, "added", form.Order)
		assert.Equal(t, "2022-01-02T13:04:05-01:00", form.Added.Format(time.RFC3339))
		assert.Equal(t, "2022-01-02T14:04:05Z", form.Added.UTC().Format(time.RFC3339))
		assert.Equal(t, "2001-01-02T17:04:05Z", form.Updated.Format(time.RFC3339))
		assert.Equal(t, 100, form.Count)
		assert.Equal(t, 0, form.Offset)
	})
}

func TestNewSearchPhotos(t *testing.T) {
	r := NewSearchPhotos("cat")
	assert.IsType(t, SearchPhotos{}, r)
}

func TestSearchPhotos_Serialize(t *testing.T) {
	form := SearchPhotos{
		Query:   "foo BAR",
		Private: true,
		Photo:   false,
		Lat:     1.5,
		Lng:     -10.33333,
		Year:    "2002",
		Chroma:  1,
		Diff:    424242,
		Before:  time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC),
	}

	result := form.Serialize()

	t.Logf("SERIALIZED: %s", result)

	assert.IsType(t, "string", result)
}

func TestSearchPhotos_SerializeAll(t *testing.T) {
	form := SearchPhotos{
		Query:   "foo BAR",
		Private: true,
		Photo:   false,
		Lat:     1.5,
		Lng:     -10.33333,
		Year:    "2002|2003",
		Chroma:  1,
		Diff:    424242,
		Before:  time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC),
	}

	result := form.SerializeAll()

	t.Logf("SERIALIZED: %s", result)

	assert.IsType(t, "string", result)
}

func TestSearchPhotos_Filter(t *testing.T) {
	t.Run("WithScope", func(t *testing.T) {
		f := &SearchPhotos{Query: "album:cat filter:\"name:foo.jpg\" s:ariqwb43p5dh9h13 search-string", Scope: "ariqwb43p5dh2244", Filter: "name:foo.jpg album:ariqwb43p5dh5555 q:foo uid:priqwb43p5dh7777"}

		err := f.ParseQueryString()

		t.Logf("WithScope: %+v\n", f)

		assert.ErrorContains(t, err, "unknown filter: s")
		assert.Equal(t, "search-string", f.Query)
		assert.Equal(t, "ariqwb43p5dh2244", f.Scope)
		assert.Equal(t, "name:foo.jpg album:ariqwb43p5dh5555 q:foo uid:priqwb43p5dh7777", f.Filter)
		assert.Equal(t, "", f.Name)
		assert.Equal(t, "", f.UID)
		assert.Equal(t, "cat", f.Album)
	})
	t.Run("ScopeInQuery", func(t *testing.T) {
		f := &SearchPhotos{Query: "album:cat filter:\"name:foo.jpg\" s:ariqwb43p5dh9h13 search-string", Filter: "name:foo.jpg album:ariqwb43p5dh5555 q:foo uid:priqwb43p5dh7777"}

		err := f.ParseQueryString()

		t.Logf("ScopeInQuery: %+v\n", f)

		assert.ErrorContains(t, err, "unknown filter: s")
		assert.Equal(t, "search-string", f.Query)
		assert.Equal(t, "", f.Scope)
		assert.Equal(t, "name:foo.jpg album:ariqwb43p5dh5555 q:foo uid:priqwb43p5dh7777", f.Filter)
		assert.Equal(t, "", f.Name)
		assert.Equal(t, "", f.UID)
		assert.Equal(t, "cat", f.Album)
	})
	t.Run("NoScope", func(t *testing.T) {
		f := &SearchPhotos{Query: "album:cat search-string", Filter: "name:foo.jpg album:ariqwb43p5dh5555 q:foo uid:priqwb43p5dh7777"}

		err := f.ParseQueryString()

		t.Logf("ScopeInQuery: %+v\n", f)

		assert.NoError(t, err)
		assert.Equal(t, "foo", f.Query)
		assert.Equal(t, "", f.Scope)
		assert.Equal(t, "name:foo.jpg album:ariqwb43p5dh5555 q:foo uid:priqwb43p5dh7777", f.Filter)
		assert.Equal(t, "foo", f.Name)
		assert.Equal(t, "priqwb43p5dh7777", f.UID)
		assert.Equal(t, "ariqwb43p5dh5555", f.Album)
	})
}

func TestSearchPhotos_Unserialize(t *testing.T) {
	t.Run("Filter", func(t *testing.T) {
		f := &SearchPhotos{Query: "bar album:ariqwb43p5dh9999 uid:priqwb43p5dh4321 albums:baz s:ariqwb43p5dh1122 search-string", Scope: "ariqwb43p5dh2244"}

		if err := Unserialize(f, "name:foo.jpg album:ariqwb43p5dh5555 q:foo uid:priqwb43p5dh7777"); err != nil {
			t.Fatal(err)
		}

		t.Logf("UnserializeFilter: %+v\n", f)

		assert.Equal(t, "foo", f.Query)
		assert.Equal(t, "ariqwb43p5dh2244", f.Scope)
		assert.Equal(t, "", f.Filter)
		assert.Equal(t, "foo.jpg", f.Name)
		assert.Equal(t, "priqwb43p5dh7777", f.UID)
		assert.Equal(t, "ariqwb43p5dh5555", f.Album)
	})
}

func TestSearchPhotos_FindUidOnly(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		f := &SearchPhotos{UID: "priqwb43p5dh7777"}

		assert.True(t, f.FindUidOnly())
	})
	t.Run("False", func(t *testing.T) {
		f := &SearchPhotos{Query: "label:cat", UID: "priqwb43p5dh7777"}

		assert.False(t, f.FindUidOnly())
	})
}
