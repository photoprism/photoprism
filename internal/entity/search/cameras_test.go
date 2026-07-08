package search

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func TestCameras(t *testing.T) {
	t.Run("SearchWithQuery", func(t *testing.T) {
		query := form.NewCameraSearch("q:Canon")
		query.Count = 1005
		result, err := Cameras(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(result))

		for _, r := range result {
			assert.IsType(t, Camera{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.CameraName)
			assert.NotEmpty(t, r.CameraSlug)
			assert.True(t, strings.Contains(strings.ToLower(r.CameraName), "canon") || strings.Contains(strings.ToLower(r.CameraMake), "canon") || strings.Contains(strings.ToLower(r.CameraModel), "canon"))

			if fix, ok := entity.CameraFixtures[r.CameraSlug]; ok {
				assert.Equal(t, fix.CameraName, r.CameraName)
				assert.Equal(t, fix.CameraSlug, r.CameraSlug)
				assert.Equal(t, fix.CameraMake, r.CameraMake)
				assert.Equal(t, fix.CameraModel, r.CameraModel)
			}
		}
	})
	t.Run("SearchForString", func(t *testing.T) {
		query := form.NewCameraSearch("Q:EOS 5D")
		query.Count = 1005
		result, err := Cameras(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(result))

		for _, r := range result {
			assert.IsType(t, Camera{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.CameraName)
			assert.NotEmpty(t, r.CameraSlug)
			assert.NotEmpty(t, r.CameraModel)
		}
	})
	t.Run("SearchForNoMake", func(t *testing.T) {
		query := form.NewCameraSearch("NoMake:true")
		query.Count = 15
		result, err := Cameras(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(result))

		for _, r := range result {
			assert.IsType(t, Camera{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.CameraName)
			assert.NotEmpty(t, r.CameraSlug)
			assert.Empty(t, r.CameraMake)
		}
	})
	t.Run("SearchWithEmptyQuery", func(t *testing.T) {
		query := form.NewCameraSearch("")
		result, err := Cameras(query)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 6, len(result))
	})
	t.Run("SearchWithReverse", func(t *testing.T) {
		query := form.NewCameraSearch("")
		query.Reverse = true
		result, err := Cameras(query)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 6, len(result))
	})
	t.Run("SearchForNoMakeAndQueryNoResults", func(t *testing.T) {
		query := form.NewCameraSearch("Canon")
		query.NoMake = true
		result, err := Cameras(query)

		assert.NoError(t, err)
		assert.Empty(t, result)
	})
	t.Run("SearchWithInvalidQueryString", func(t *testing.T) {
		query := form.NewCameraSearch("xxx:bla")
		result, err := Cameras(query)

		assert.Error(t, err, "unknown filter")
		assert.Empty(t, result)
	})
	t.Run("SearchForId", func(t *testing.T) {
		f := form.SearchCameras{
			ID:    "1000000|1000001",
			Count: 0,
		}

		result, err := Cameras(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, result, 2)
	})
}
