package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosGeoFilterAlbums(t *testing.T) {
	t.Run("Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Berlin 2019"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Berlin 2019"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Pet*"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Pet*"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* pipe Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Pet*|Berlin 2019"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Pet*|Berlin 2019"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		t.Log(photos)
		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* whitespace pipe whitespace Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Pet* | Berlin 2019"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Pet* | Berlin 2019"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* or Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Pet* or Berlin 2019"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Pet* or Berlin 2019"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* OR Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Pet* OR Berlin 2019"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Pet* OR Berlin 2019"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)

	})
	t.Run("Pet* Ampersand Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Pet*&Berlin 2019"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Pet*&Berlin 2019"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* whitespace Ampersand whitespace Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Pet* & Berlin 2019"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Pet* & Berlin 2019"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* and Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Pet* and Berlin 2019"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Pet* and Berlin 2019"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* AND Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Pet* AND Berlin 2019"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Pet* AND Berlin 2019"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "%gold"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)

		var geo form.SearchPhotosGeo

		geo.Albums = "%gold"

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

		f.Albums = "I love % dog"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)

		var geo form.SearchPhotosGeo

		geo.Albums = "I love % dog"

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

		f.Albums = "sale%"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "sale%"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "&IlikeFood"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)

		var geo form.SearchPhotosGeo

		geo.Albums = "&IlikeFood"

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

		f.Albums = "Pets & Dogs"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)

		var geo form.SearchPhotosGeo

		geo.Albums = "Pets & Dogs"

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

		f.Albums = "Light&"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Light&"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "'Family"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)

		var geo form.SearchPhotosGeo

		geo.Albums = "'Family"

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

		f.Albums = "Father's Day"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Father's Day"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Ice Cream'"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Ice Cream'"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "*Forrest"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)

		var geo form.SearchPhotosGeo

		geo.Albums = "*Forrest"

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

		f.Albums = "My*Kids"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "My*Kids"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Yoga***"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Yoga***"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "|Banana"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)

		var geo form.SearchPhotosGeo

		geo.Albums = "|Banana"

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

		f.Albums = "Red|Green"
		f.Primary = true
		f.Geo = "yes"

		photos, count, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		if len(photos) != 1 {
			t.Logf("excactly one result expected, but %d photos with %d files found", len(photos), count)
			t.Logf("query results: %#v", photos)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Red|Green"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Blue|"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Blue|"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "345 Shirt"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "345 Shirt"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Color555 Blue"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Color555 Blue"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Route 66"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Route 66"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("AndSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Route 66 & Father's Day"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Route 66 & Father's Day"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("OrSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Route 66 | Father's Day"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Route 66 | Father's Day"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("AndSearch2", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Red|Green & Father's Day"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Red|Green & Father's Day"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("OrSearch2", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Red|Green | Father's Day"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Red|Green | Father's Day"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("AndSearch3", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Light& & Red|Green"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Light& & Red|Green"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("OrSearch3", func(t *testing.T) {
		var f form.SearchPhotos

		f.Albums = "Red|Green | Light&"
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Albums = "Red|Green | Light&"

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
}

func TestPhotosGeoQueryAlbums(t *testing.T) {
	t.Run("Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Berlin 2019\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Berlin 2019\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)

	})
	t.Run("Pet*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Pet*\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Pet*\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* pipe Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Pet*|Berlin 2019\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Pet*|Berlin 2019\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* whitespace pipe whitespace Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Pet* | Berlin 2019\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Pet* | Berlin 2019\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* or Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Pet* or Berlin 2019\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Pet* or Berlin 2019\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* OR Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Pet* OR Berlin 2019\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Pet* OR Berlin 2019\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* Ampersand Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Pet*&Berlin 2019\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Pet*&Berlin 2019\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* whitespace Ampersand whitespace Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Pet* & Berlin 2019\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Pet* & Berlin 2019\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* and Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Pet* and Berlin 2019\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Pet* and Berlin 2019\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("Pet* AND Berlin 2019", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Pet* AND Berlin 2019\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Pet* AND Berlin 2019\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"%gold\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"%gold\""

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

		f.Query = "albums:\"I love % dog\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"I love % dog\""

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

		f.Query = "albums:\"sale%\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"sale%\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"&IlikeFood\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"&IlikeFood\""

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

		f.Query = "albums:\"Pets & Dogs\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Pets & Dogs\""

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

		f.Query = "albums:\"Light&\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Light&\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"'Family\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"'Family\""

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

		f.Query = "albums:\"Father's Day\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Father's Day\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Ice Cream'\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Ice Cream'\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"*Forrest\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"*Forrest\""

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

		f.Query = "albums:\"My*Kids\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"My*Kids\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Yoga***\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Yoga***\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"|Banana\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"|Banana\""

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

		f.Query = "albums:\"Red|Green\""
		f.Primary = true
		f.Geo = "yes"

		photos, count, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		if len(photos) != 1 {
			t.Logf("excactly one result expected, but %d photos with %d files found", len(photos), count)
			t.Logf("query results: %#v", photos)
		}

		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Red|Green\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Blue|\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Blue|\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"345 Shirt\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"345 Shirt\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Color555 Blue\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Color555 Blue\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Route 66\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Route 66\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("AndSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Route 66 & Father's Day\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Route 66 & Father's Day\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("OrSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Route 66 | Father's Day\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Route 66 | Father's Day\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("AndSearch2", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Red|Green & Father's Day\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Red|Green & Father's Day\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("OrSearch2", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Red|Green | Father's Day\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Red|Green | Father's Day\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("AndSearch3", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Light& & Red|Green\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Light& & Red|Green\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
	t.Run("OrSearch3", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "albums:\"Red|Green | Light&\""
		f.Primary = true
		f.Geo = "yes"

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		var geo form.SearchPhotosGeo

		geo.Query = "albums:\"Red|Green | Light&\""

		// Parse query string and filter.
		if err = geo.ParseQueryString(); err != nil {
			t.Fatal(err)
		}

		geophotos, err2 := PhotosGeo(geo)

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.Equal(t, len(photos), len(geophotos))
		assert.Equal(t, photos[0].PhotoUID, geophotos[0].PhotoUID)
	})
}
