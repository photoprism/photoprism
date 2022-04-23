package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestPhotosFilterFaces(t *testing.T) {
	var f0 form.SearchPhotos

	f0.Primary = true

	photos0, _, _ := Photos(f0)
	t.Run("yes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "yes"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 6)
	})
	t.Run("1", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "1"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 6)
	})
	t.Run("2", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "2"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 3)
	})
	t.Run("5", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "4"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "%gold"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	//TODO random result
	/*t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "I love % dog"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})*/
	//TODO random result
	/*t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "sale%"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})*/
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "&IlikeFood"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	//TODO random result
	/*t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "Pets & Dogs"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})*/
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "Light&"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "'Family"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	//TODO random result
	/*t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "Father's faces"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
	})*/
	//TODO random result
	/*t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "Ice Cream'"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})*/
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "*Forrest"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "My*Kids"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	//TODO random result
	/*t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "Yoga***"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})*/
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "|Banana"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "Red|Green"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "Blue|"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "345 Shirt"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	//TODO random result
	/*t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "faces555 Blue"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})*/
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Faces = "Route 66"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
}

func TestPhotosQueryFaces(t *testing.T) {
	var f0 form.SearchPhotos

	f0.Primary = true

	photos0, _, _ := Photos(f0)

	t.Run("yes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:yes"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 6)
	})
	t.Run("1", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:1"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 6)
	})
	t.Run("2", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:2"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 3)
	})
	t.Run("5", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:5"
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"%gold\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	//TODO random result
	/*t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"I love % dog\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})*/
	//TODO random result
	/*t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"sale%\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})*/
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"&IlikeFood\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	//TODO random result
	/*t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"Pets & Dogs\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})*/
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"Light&\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"'Family\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	//TODO random result
	/*t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"Father's faces\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
	})*/
	//TODO random result
	/*t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"Ice Cream'\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})*/
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"*Forrest\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"My*Kids\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	//TODO random result
	/*t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"Yoga***\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})*/
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"|Banana\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"Red|Green\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), len(photos0))
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"Blue|\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"345 Shirt\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
	//TODO random result
	/*t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"faces555 Blue\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})*/
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "faces:\"Route 66\""
		f.Primary = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), len(photos0))
	})
}
