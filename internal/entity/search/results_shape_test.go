package search

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// shapeNoMatch is a search term that matches no fixture.
const shapeNoMatch = "shapenomatchxyz"

// assertEmptyJsonList checks that a search returned no rows that encode as an empty JSON list.
func assertEmptyJsonList(t *testing.T, results any, err error) {
	t.Helper()

	require.NoError(t, err)

	data, err := json.Marshal(results)
	require.NoError(t, err)
	assert.Equal(t, "[]", string(data))
}

// TestEmptyResults_Json pins that searches without matches encode as [] rather than null.
func TestEmptyResults_Json(t *testing.T) {
	t.Run("Photos", func(t *testing.T) {
		results, _, err := Photos(form.SearchPhotos{Query: shapeNoMatch, Count: 10})
		assertEmptyJsonList(t, results, err)
	})
	t.Run("PhotosViewerResults", func(t *testing.T) {
		results, _, err := PhotosViewerResults(form.SearchPhotos{Query: shapeNoMatch, Count: 10}, "/content", "/api/v1", "preview", "download")
		assertEmptyJsonList(t, results, err)
	})
	t.Run("PhotosGeo", func(t *testing.T) {
		results, err := PhotosGeo(form.SearchPhotosGeo{Query: shapeNoMatch, Count: 10})
		assertEmptyJsonList(t, results, err)
	})
	t.Run("Albums", func(t *testing.T) {
		results, err := Albums(form.SearchAlbums{Query: shapeNoMatch, Count: 10})
		assertEmptyJsonList(t, results, err)
	})
	t.Run("AlbumPhotos", func(t *testing.T) {
		a := entity.NewAlbum("Shape Empty Album", entity.AlbumManual)
		require.NoError(t, a.Create())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(a).Error })

		results, err := AlbumPhotos(*a, 10, false)
		assertEmptyJsonList(t, results, err)
	})
	t.Run("Labels", func(t *testing.T) {
		results, err := Labels(form.SearchLabels{Query: shapeNoMatch, Count: 10})
		assertEmptyJsonList(t, results, err)
	})
	t.Run("Cameras", func(t *testing.T) {
		results, err := Cameras(form.SearchCameras{Query: shapeNoMatch, Count: 10})
		assertEmptyJsonList(t, results, err)
	})
	t.Run("Lenses", func(t *testing.T) {
		results, err := Lenses(form.SearchLenses{Query: shapeNoMatch, Count: 10})
		assertEmptyJsonList(t, results, err)
	})
	t.Run("Faces", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{UID: "FACENOMATCHSHAPEXYZ", Count: 10})
		assertEmptyJsonList(t, results, err)
	})
	t.Run("Subjects", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Query: shapeNoMatch, Count: 10})
		assertEmptyJsonList(t, results, err)
	})
	t.Run("Users", func(t *testing.T) {
		results, err := Users(form.SearchUsers{Query: shapeNoMatch, Count: 10})
		assertEmptyJsonList(t, results, err)
	})
	t.Run("Sessions", func(t *testing.T) {
		results, err := Sessions(form.SearchSessions{Query: shapeNoMatch, Count: 10})
		assertEmptyJsonList(t, results, err)
	})
	t.Run("Accounts", func(t *testing.T) {
		results, err := Accounts(form.SearchServices{Query: shapeNoMatch, Count: 10})
		assertEmptyJsonList(t, results, err)
	})
}
