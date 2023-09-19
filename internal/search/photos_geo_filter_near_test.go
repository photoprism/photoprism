package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosGeoFilterNear(t *testing.T) {
	t.Run("pt9jtdre2lvl0y24", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "pt9jtdre2lvl0y24"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 8)
	})
	t.Run("pr2xu7myk7wrbk30", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "pr2xu7myk7wrbk30"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 26)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "%gold"
		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "record not found")
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "I love % dog"
		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "record not found")
	})
	//TODO error
	/*t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "sale%"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "&IlikeFood"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Pets & Dogs"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Light&"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "'Family"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Father's type"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Ice Cream'"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "*Forrest"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "My*Kids"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Yoga***"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "|Banana"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Red|Green"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Blue|"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "345 Shirt"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "type555 Blue"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Route 66"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})*/
}

func TestPhotosGeoQueryNear(t *testing.T) {
	t.Run("pt9jtdre2lvl0y24", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:pt9jtdre2lvl0y24"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 8)
	})
	t.Run("pr2xu7myk7wrbk30", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:pr2xu7myk7wrbk30"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 26)
	})
	//TODO error
	/*t.Run("pt9jtdre2lvl0y24 pipe pr2xu7myk7wrbk30", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:pt9jtdre2lvl0y24|pr2xu7myk7wrbk30"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"%gold\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"I love % dog\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"sale%\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"&IlikeFood\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Pets & Dogs\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Light&\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"'Family\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Father's type\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Ice Cream'\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"*Forrest\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"My*Kids\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Yoga***\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"|Banana\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Red|Green\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Blue|\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"345 Shirt\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"type555 Blue\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Route 66\""

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})*/
}
