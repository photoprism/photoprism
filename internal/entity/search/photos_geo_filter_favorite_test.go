package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestPhotosGeoQueryFavorite(t *testing.T) {
	var f0 form.SearchPhotos

	f0.Query = "favorite:true"
	f0.Primary = true
	f0.Geo = "yes"

	// Parse query string and filter.
	if err := f0.ParseQueryString(); err != nil {
		t.Fatal(err)
	}

	photos0, _, err := Photos(f0)

	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(photos0), 6)

	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"%gold\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"%gold\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"I love % dog\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"I love % dog\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"sale%\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"sale%\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"&IlikeFood\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"&IlikeFood\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"Pets & Dogs\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"Pets & Dogs\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"Light&\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"Light&\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"'Family\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"'Family\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		// Note: If the string in favorite starts with f/F, the txt package will assume it means false,
		f.Query = "favorite:\"Mother's Day\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"Mother's Day\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"Ice Cream'\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"Ice Cream'\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"*Forrest\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"*Forrest\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"My*Kids\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"My*Kids\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"Yoga***\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"Yoga***\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"|Banana\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"|Banana\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"Red|Green\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"Red|Green\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"Blue|\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"Blue|\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"345 Shirt\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"345 Shirt\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"Color555 Blue\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"Color555 Blue\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"Route 66\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"Route 66\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("AndSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"Route 66 & Father's Day\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"Route 66 & Father's Day\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
	t.Run("OrSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "favorite:\"Route %66 | *Father's Day\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
		var geo form.SearchPhotosGeo

		geo.Query = "favorite:\"Route %66 | *Father's Day\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
	})
}
