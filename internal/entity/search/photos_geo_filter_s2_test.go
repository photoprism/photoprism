package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosGeoFilterS2(t *testing.T) {
	t.Run("1ef744d1e283", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "1ef744d1e283"
		f.Dist = 2

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 3)
	})
	t.Run("85d1ea7d382c", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "85d1ea7d382c"
		f.Dist = 2

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 8)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "%gold"
		f.Dist = 2

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "I love % dog"
		f.Dist = 2

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "sale%"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "&IlikeFood"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "Pets & Dogs"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "Light&"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "'Family"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "Father's type"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "Ice Cream'"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "*Forrest"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "My*Kids"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "Yoga***"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "|Banana"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "Red|Green"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "Blue|"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(photos))
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "345 Shirt"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "type555 Blue"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.S2 = "Route 66"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
}

func TestPhotosGeoQueryS2(t *testing.T) {
	t.Run("s2:1ef744d1e283", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:1ef744d1e283"
		f.Dist = 2

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 3, len(photos))
	})
	t.Run("s2:85d1ea7d382c", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:85d1ea7d382c"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 8, len(photos))
	})
	t.Run("85d1ea7d382c pipe 1ef744d1e283", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:85d1ea7d382c|1ef744d1e283"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"%gold\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"I love % dog\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"sale%\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"&IlikeFood\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"Pets & Dogs\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"Light&\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"'Family\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"Father's type\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"Ice Cream'\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"*Forrest\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"My*Kids\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"Yoga***\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"|Banana\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"Red|Green\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"Blue|\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"345 Shirt\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"type555 Blue\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "s2:\"Route 66\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
}
