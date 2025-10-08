package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosGeoFilterNear(t *testing.T) {
	t.Run("Ps6sg6be2lvl0y24", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "ps6sg6be2lvl0y24"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 9)
	})
	t.Run("Ps6sg6byk7wrbk30", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "ps6sg6byk7wrbk30"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 26)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "%gold"
		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "I love % dog"
		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "sale%"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "&IlikeFood"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Pets & Dogs"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Light&"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "'Family"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Father's type"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Ice Cream'"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "*Forrest"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "My*Kids"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Yoga***"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "|Banana"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Red|Green"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Blue|"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "345 Shirt"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "type555 Blue"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Near = "Route 66"

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
}

func TestPhotosGeoQueryNear(t *testing.T) {
	t.Run("Ps6sg6be2lvl0y24", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:ps6sg6be2lvl0y24"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 9)
	})
	t.Run("Ps6sg6byk7wrbk30", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:ps6sg6byk7wrbk30"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 26)
	})
	t.Run("Ps6sg6be2lvl0y24PipePs6sg6byk7wrbk30", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:ps6sg6be2lvl0y24|ps6sg6byk7wrbk30"

		photos, err := PhotosGeo(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 35)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"%gold\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"I love % dog\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"sale%\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"&IlikeFood\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Pets & Dogs\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Light&\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"'Family\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Father's type\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Ice Cream'\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"*Forrest\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"My*Kids\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Yoga***\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"|Banana\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Red|Green\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Blue|\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"345 Shirt\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"type555 Blue\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotosGeo

		f.Query = "near:\"Route 66\""

		_, err := PhotosGeo(f)

		assert.Equal(t, err.Error(), "Not found")
	})
}
